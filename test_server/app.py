import functools
import hashlib
import traceback

from flask import Flask, make_response, request
from werkzeug.exceptions import HTTPException

import crypto
import encoder
from encoder import Encoding


@functools.lru_cache()
def AES_key():
    # if secret is set in app config, used to produce AES key
    # otherwise, just generate random one
    secret = app.config.get("SECRET")
    if secret is None:
        return crypto.random_bytes(16)
    else:
        return hashlib.md5(secret.encode()).digest()


app = Flask(__name__)

def get_encoding(request):
    # get encoding (defaults to Base64 if not specified)
    encoding = request.values.get("enc", None)
    if encoding is None:
        encoding = Encoding.B64.name
    else:
        encoding = encoding.upper()

    return encoding


# encrypts the plaintext
@app.route("/encrypt", methods=["GET", "POST"])
def route_encrypt():
    # get plaintext to encrypt
    plain = request.values.get("plain", None)
    if plain is None:
        raise ValueError(
            "Pass data to encrypt using 'plain' parameter in URL or POST data"
        )

    # encrypt the data (encoded to bytes)
    cipher = crypto.encrypt(plain.encode(), AES_key())

    # get encoding
    encoding = get_encoding(request)

    # encode encrypted chunk
    encoded_cipher = encoder.encode(data=cipher, encoding=Encoding[encoding])

    # answer
    return encoded_cipher, 200


# decrypts the cipher
@app.route("/decrypt", methods=["GET", "POST"])
def route_decrypt():
    # get ciphertext
    encoded_cipher = request.values.get("cipher", None)
    if encoded_cipher is None:
        raise ValueError(
            "Pass encoded chipher to decrypt using 'cipher' parameter in URL or POST data"
        )

    # get encoding
    encoding = get_encoding(request)

    # decode cipher into bytes
    cipher = encoder.decode(data=encoded_cipher, encoding=Encoding[encoding])

    # decrypt cipher into plaintext
    plain = crypto.decrypt(cipher, AES_key())

    # answer
    return plain, 200


@app.route("/health")
def health():
    return "OK", 200


# this is what makes the server vulnerable to padding oracle
# it just talks too much about errors
# NOTE: to test Padding Oracle detection, every exception's trace is printed
# (not just IncorrectPadding)
@app.errorhandler(Exception)
def handle_incorrect_padding(exc):
    # pass through HTTP errors
    if isinstance(exc, HTTPException):
        return exc

    # log exception
    # app.logger.exception(exc)

    if app.config.get("VULNERABLE"):
        # vulnerable response
        response = make_response(traceback.format_exc(), 500)
        response.headers["content-type"] = "text/plain"
        return response
    else:
        # non-vulnerable response
        return "Internal server error", 500
