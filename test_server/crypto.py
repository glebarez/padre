import random

from Crypto.Cipher import AES
from Crypto.Util import Padding


def random_bytes(length: int) -> bytes:
    out = []
    for _ in range(length):
        out.append(random.randint(0, 0xFF))
    return bytes(out)


class IncorrectPadding(Exception):
    def __init__(self):
        super(IncorrectPadding, self).__init__("Incorrect Padding")


def encrypt(data: bytes, key: bytes) -> bytes:
    # pad data
    data = Padding.pad(data, 16)

    # new encryptor
    encryptor = AES.new(key, AES.MODE_CBC)

    # return IV + cipher
    return encryptor.iv + encryptor.encrypt(data)


def decrypt(data: bytes, key: bytes) -> str:
    # tell IV from cipher
    iv, data = data[:16], data[16:]

    # fresh encryptor with IV provided
    encryptor = AES.new(key, AES.MODE_CBC, iv)

    # decrypt
    plain = encryptor.decrypt(data)

    # unpad, decode, return
    try:
        return Padding.unpad(plain, 16)
    except ValueError:
        raise IncorrectPadding()
