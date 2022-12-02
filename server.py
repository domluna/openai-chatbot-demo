"""Make some requests to OpenAI's chatbot"""

import os
import time

import flask

from flask import g

from playwright.sync_api import sync_playwright

APP = flask.Flask(__name__)
PLAY = sync_playwright().start()
BROWSER = PLAY.chromium.launch_persistent_context(
    user_data_dir="/tmp/playwright",
    headless=False,
)
PAGE = BROWSER.new_page()


def get_input_box():
    """Get the child textarea of `PromptTextarea__TextareaWrapper`"""

    p = PAGE.query_selector("div[class*='PromptTextarea__TextareaWrapper']")
    if p is None:
        return False
    return p.query_selector("textarea")


def is_logged_in():
    # See if we have a textarea with data-id="root"
    return get_input_box() is not None


def send_message(message):
    # Send the message
    box = get_input_box()
    box.click()
    box.fill(message)
    box.press("Enter")


def get_last_message():
    """Get the latest message"""
    page_elements = PAGE.query_selector_all("div[class*='ConversationItem__Message']")
    last_element = page_elements[-1]
    return last_element.inner_text()


def done_generating_response():
    el = PAGE.query_selector(
        "div[class*='PromptTextarea__LastItemActions-sc-4snkpf-3 gRmLdg']"
    )
    button = el.query_selector("button")
    if button is None:
        return False
    return True


def reset_chat():
    page_elements = PAGE.query_selector_all(
        "a[class*='Navigation__NavMenuItem-sc-rtsy24-0 dWwJhv']"
    )
    reset_box = page_elements[0]
    reset_box.click()
    print("Reset chat")
    return None


def redo_response():
    el = PAGE.query_selector(
        "div[class*='PromptTextarea__LastItemActions-sc-4snkpf-3 gRmLdg']"
    )
    # select the button from el
    button = el.query_selector("button")
    button.click()
    return None


@APP.route("/chat", methods=["GET"])
def chat():
    message = flask.request.args.get("q")
    print("Sending message: ", message)
    if message == "RESET":
        reset_chat()
        return "Reset chat"
    elif message == "REDO":
        print("Redoing response")
        redo_response()
    else:
        send_message(message)

    i = 0
    while True:
        if done_generating_response():
            print("DONE GENERATING RESPONSE")
            break
        else:
            time.sleep(1)
            i += 1

        if i > 30:
            print("Timeout !!! Grabbing response as is")
            break
    response = get_last_message()
    print("Response: ", response)
    return response


def start_browser():
    PAGE.goto("https://chat.openai.com/")
    if not is_logged_in():
        print("Please log in to OpenAI Chat")
        print("Press enter when you're done")
        input()
    else:
        print("Logged in")
        APP.run(port=5001, threaded=False)


if __name__ == "__main__":
    start_browser()
