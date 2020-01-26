# A simple flask web server just to test things out
debug = False

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
app.config['DEBUG'] = False
secret = 'Some really secret key'
key = hashlib.md5(secret.encode()).digest()

def decode(data):
    # x = data.replace('~', '=').replace('!', '/').replace('-', '+')
    x = ba.a2b_base64(data)
    return x

def encode(data):
    x = ba.b2a_base64(data).decode()[:-1]
    # x = x.replace('=', '~').replace('/', '!').replace('+', '-')
    return x

@app.route('/encrypt')
def route_encrypt():
    # get plaintext
    plain = request.args.get('plain', None)
    if not plain:
        return 'No plain', 500

    # encrypt
    cipher = cry.encrypt(plain, key)

    # return
    return encode(cipher), 200

@app.route('/encryptb64')
def route_encryptb64():
    # get plaintext
    plain = request.args.get('plain', None)
    if not plain:
        return 'No plain', 
    
    # base64 to bytes
    plain = ba.a2b_base64(plain)

    # encrypt
    cipher = cry.encrypt(plain, key, binary = True)

    # return
    return encode(cipher), 200

@app.route('/decrypt', methods = ['GET','POST'])
def route_decrypt():
    # artifical sleep to imitate real-world web server
    time.sleep(random.random()/4 + .1)

    # get cipher
    if request.method == 'GET':
        cipher = request.args.get('cipher',None)
    else:
        cipher = request.form.get('cipher',None)

    if not cipher:
        return 'No cipher', 500

    try:
        cipher = decode(cipher)
        plain = cry.decrypt(cipher, key)
        return plain, 200
    except Exception:
        response = make_response(traceback.format_exc(),500)
        response.headers["content-type"] = "text/plain"
        return response


if __name__ == '__main__':
    if not debug:
        WSGIServer((
            "127.0.0.1", # str(HOST)
            5000,  # int(PORT)
        ), app.wsgi_app).serve_forever()
    else:
        app.run()
