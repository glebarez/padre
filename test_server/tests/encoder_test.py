import pytest

from crypto import random_bytes
from encoder import Encoding, decode, encode


@pytest.mark.parametrize("value", [random_bytes(i) for i in range(10)], ids=len)
@pytest.mark.parametrize("encoding", list(Encoding))
def test_encoding_decoding(value, encoding):
    encoded = encode(value, encoding)
    decoded = decode(encoded, encoding)
    assert decoded == value


def test_unknown_encoding():
    with pytest.raises(RuntimeError):
        encode(b"", -1)

    with pytest.raises(RuntimeError):
        decode("", -1)
