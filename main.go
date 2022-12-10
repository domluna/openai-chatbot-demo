package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	sqlite3 "github.com/mattn/go-sqlite3"
)

type Chat struct {
	ID        int       `json:"id" db:"id"`
	Done      bool      `json:"done" db:"done"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Message struct {
	ID        int       `json:"id" db:"id"`
	Message   string    `json:"message" db:"message"`
	Response  string    `json:"response" db:"response"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ChatID    int       `json:"chat_id" db:"chat_id"`
}

const (
	BaseURL = "http://localhost:5001/chat"
	DBDir  = ".chatgpt"
	DBName  = "chatgpt.db"
)

func endChat(db *sql.DB, chatID int64) error {
	url := fmt.Sprintf("%s?q=RESET", BaseURL)
	_, err := http.Get(url)
	if err != nil {
		return err
	}

	// Mark the chat as done
	_, err = db.Exec("UPDATE chats SET done = true WHERE id = ?", chatID)
	if err != nil {
		return err
	}
	return nil
}

func askChat(db *sql.DB, chatID int64, question string) (string, error) {
	// Ask the server for a question
	url := fmt.Sprintf("%s?q=%s", BaseURL, url.QueryEscape(question))
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	response := string(body)

	// Save the question and response the messages tables
	_, err = db.Exec("INSERT INTO messages (message, response, chat_id) VALUES (?, ?, ?)", question, response, chatID)

	return response, nil
}

func newChat(db *sql.DB) (int64, error) {
	// If there is no chat, create one
	stmt, err := db.Prepare(`INSERT INTO chats DEFAULT VALUES`)
	if err != nil {
		return 0, err
	}
	// Execute the INSERT statement and insert the data
	res, err := stmt.Exec()
	if err != nil {
		return 0, err
	}

	// Retrieve the ID of the inserted record
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func renewChat(db *sql.DB, chatID int64) (int64, error) {
	err := endChat(db, chatID)
	if err != nil {
		log.Fatal(err)
		return 0, err
	}
	chatID, err = newChat(db)
	if err != nil {
		log.Fatal(err)
		return 0, err
	}

	return chatID, nil
}

func main() {
	home, err := os.UserHomeDir()
	dir := fmt.Sprintf("%s/%s", home, DBDir)
	// Check if the directory already exists.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Create the directory.
		err = os.Mkdir(dir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory: %s\n", err)
			return
		}
		fmt.Println("Directory created successfully.")
	} 

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s/%s", dir, DBName))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	// Create the table if it doesn't already exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS chats (
			id INTEGER PRIMARY KEY,
			done BOOLEAN NOT NULL DEFAULT false,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		if err == sqlite3.ErrError {
			// Table already exists, ignore the error
			return
		}
		log.Fatal(err)
		return
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY,
			message TEXT NOT NULL,
			response TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			chat_id INTEGER NOT NULL,

			FOREIGN KEY(chat_id) REFERENCES chats(id) 
		)
	`)
	if err != nil {
		if err == sqlite3.ErrError {
			// Table already exists, ignore the error
			return
		}
		log.Fatal(err)
		return
	}

	// Parse the command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <command> [<question>]")
		os.Exit(1)
	}
	command := os.Args[1]
	question := ""
	if len(os.Args) >= 3 {
		question = os.Args[2]
	}

	// Get the current most recent unfinished chat id
	isFreshChat := false
	var chatID int64
	err = db.QueryRow("SELECT id FROM chats WHERE not done ORDER BY id DESC LIMIT 1").Scan(&chatID)
	if err != nil {
		// If there is no chat, create one
		chatID, err = newChat(db)
		if err != nil {
			log.Fatal(err)
			return
		}
		isFreshChat = true
	}

	fmt.Printf("QUESTION\n\n%s\n\n", question)

	// Make the request to the server
	if command == "cont" {
		response, err := askChat(db, chatID, question)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println("RESPONSE")
		fmt.Println(response)
	} else if command == "ask" {
		if !isFreshChat {
			chatID, err = renewChat(db, chatID)
			if err != nil {
				log.Fatal(err)
				return
			}
		}
		response, err := askChat(db, chatID, question)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Printf("RESPONSE\n\n%s\n\n", response)
	} else if command == "reset" {
		endChat(db, chatID)
		fmt.Println("Chat with ID", chatID, "ended")
	} else {
		fmt.Println("Invalid command")
		return
	}

	fmt.Println("-------------------------")
	fmt.Println("Current Chat ID:", chatID)
	fmt.Println("-------------------------")
}
