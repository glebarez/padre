# A simple flask web server just to test things out
import os
debug = "DEBUG" in os.environ

# let run as gevent if not debugging
if not debug:
    from gevent import monkey
    monkey.patch_all()
    from gevent.pywsgi import WSGIServer

from flask import Flask, request, abort, make_response
import binascii as ba
import hashlib
from Crypto.Cipher import AES
import cry
import traceback
import random, time

app = Flask(__name__)
app.config['DEBUG'] = debug
secret = 'Some really secret key'
key = hashlib.md5(secret.encode()).digest()

# decodes binary data from plaintext
def decode(data, enc = None):
    # default encoding is base64
    enc = 'b64' if enc is None else enc

    if enc == 'b64':
        x = ba.a2b_base64(data)
    elif enc == 'lhex':
        x = ba.unhexlify(data)
    else:
        raise Exception(f'Unknown encoding type: {enc}')
    return x

# encodes binary data as plaintext
def encode(data, enc = None):
    # default encoding is base64
    enc = 'b64' if enc is None else enc

    if enc == 'b64':
        x = ba.b2a_base64(data).decode()[:-1]
    elif enc == 'lhex':
        x = ba.hexlify(data).decode()
    else:
        raise Exception(f'Unknown encoding type: {enc}')
    return x

# encrypts the plaintext
@app.route('/encrypt',methods = ['GET','POST'])
def route_encrypt():
    # get plaintext
    plain = request.values.get('plain', None)
    if not plain:
        return 'No plain', 500

    # check if input is base64 encoded data
    if 'b64' in request.values:
        plain = ba.a2b_base64(plain)

    # encrypt data
    cipher = cry.encrypt(plain, key)

    # get preferred output encoding
    enc = request.values.get('enc', None)
    return encode(cipher, enc)+'\n', 200

# this route is deliberately vulnerable Padding Oracle!
@app.route('/decrypt', methods = ['GET','POST'])
def route_decrypt():
    # artifical sleep to imitate real-world web server
    time.sleep(random.random()/4 + .1)

    # get cipher
    cipher = request.values.get('cipher',None)
    if not cipher:
        return 'No cipher', 500

    # handle binary data encoding
    enc = request.values.get('enc',None)

    try:
        cipher = decode(cipher, enc)
        plain = cry.decrypt(cipher, key)
        return plain, 200
    except Exception:
        response = make_response(traceback.format_exc(),500)
        response.headers["content-type"] = "text/plain"
        return response

# main guard
if __name__ == '__main__':
    if not debug:
        WSGIServer((
            "127.0.0.1", # str(HOST)
            5000,  # int(PORT)
        ), app.wsgi_app).serve_forever()
    else:
        app.run()
