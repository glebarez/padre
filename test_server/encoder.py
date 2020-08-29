import binascii as ba
from enum import Enum, auto


class Encoding(Enum):
    B64 = auto()  # base64
    LHEX = auto()  # lowercase hex


# decodes data
def decode(data: str, encoding: Encoding) -> bytes:
    if encoding == Encoding.B64:
        x = ba.a2b_base64(data)
    elif encoding == Encoding.LHEX:
        x = ba.unhexlify(data)
    else:
        raise RuntimeError(f"Unknown encoding {encoding}")
    return x


# encodes binary data as plaintext string
def encode(data, encoding: Encoding) -> str:
    if encoding == Encoding.B64:
        x = ba.b2a_base64(data).decode()[:-1]
    elif encoding == Encoding.LHEX:
        x = ba.hexlify(data).decode()
    else:
        raise RuntimeError(f"Unknown encoding {encoding}")
    return x
