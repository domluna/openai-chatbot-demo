import json

import flask
from revChatGPT.revChatGPT import Chatbot

# read config.json
with open("config.json") as f:
    config = json.load(f)

config["Authorization"] = None
chatbot = Chatbot(config, conversation_id=None)
chatbot.reset_chat()  # Forgets conversation
chatbot.refresh_session()  # Uses the session_token to get a new bearer token

APP = flask.Flask(__name__)


@APP.route("/chat", methods=["GET"])
def chat():
    message = flask.request.args.get("q")
    print("Sending message: ", message)
    if message == "RESET":
        chatbot.reset_chat()
        return "Reset chat"
    elif message == "REFRESH SESSION":
        print("Refreshing session")
        chatbot.refresh_session()

    resp = chatbot.get_chat_response(message, output="text")
    return resp["message"]


if __name__ == "__main__":
    APP.run(port=5001, threaded=False)
