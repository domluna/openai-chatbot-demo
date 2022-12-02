"""make requests to chatbot server"""
import requests

def send_message(msg):
    url = f'http://127.0.0.1:5001/chat?q={msg}'

    # get get request to url with query parameter q as `msg` variable.
    response = requests.get(url)
    # parse the response
    output = response.text
    return output

def main():
    # start a loop to send messages to the chatbot
    # exit if Ctrl+C is pressed
    while True:
        try:
            msg = input("You:\n")
            output = send_message(msg)
            print("Response:\n", output)
        except KeyboardInterrupt:
            print("Exiting")
            break

if __name__ == "__main__":
    main()
