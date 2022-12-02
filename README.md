# OpenAI Chatbot

Modified from https://github.com/danielgross/whatsapp-gpt.

Interface for working using OpenAI Chatbot in the terminal.

Start the server with:

```
python server.py
```

Start the chatbot with:

```
python chat.py
```

You can then type questions in the chat window.

```
You:
what is 2 + 2
Response:
 2 + 2 is 4.
You:
REDO
Response:
 2 2 is a mathematical expression that equals 4.
You:
RESET
Response:
 Reset chat
You:
```

- To reset the thread type "RESET".
- To redo the last response type "REDO"

The original implementation grabbed the response after 10 seconds. This version waits until the response is complete or 30 seconds has gone by.
