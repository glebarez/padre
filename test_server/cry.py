import binascii as ba
from Crypto.Cipher import AES
from Crypto.Random import get_random_bytes
import binascii

def encrypt(data, key):
    # convert to bytes if string passed
    if isinstance(data,(str,)):
        data = data.encode()
        
    data = pad(data)     # pad for AES

    # new encryptor
    encryptor = AES.new(key, AES.MODE_CBC)

    # return IV + cipher
    return encryptor.iv + encryptor.encrypt(data)

def decrypt(data, key):
    # tell IV from cipher
    iv, data = data[:16], data[16:]

    # fresh encryptor with IV provided
    encryptor = AES.new(key, AES.MODE_CBC, iv)

    # decrypt
    plain = encryptor.decrypt(data)

    # unpad, decode, return
    return unpad(plain).decode()


def pad(e):
    l = len(e)
    p = 16-(l % 16) if (l%16) else 0
    return e + bytes([p] * p)

class IncorrectPadding(Exception):
    def __init__(self):
        super(IncorrectPadding,self).__init__("Incorrect Padding")

def unpad(data):
    # get padding byte
    padVal = data[-1]
    
    # check the padding is correct
    if len(data) < padVal:
        raise IncorrectPadding()

    for byte in data[-padVal:]:
        if byte != padVal:
            raise IncorrectPadding()

    # unpad
    return data[:-padVal]