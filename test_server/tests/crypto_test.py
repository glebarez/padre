import Crypto.Cipher.AES
import pytest

import crypto

KEY_LENGTH = 16


@pytest.fixture
def AES_key():
    return crypto.random_bytes(KEY_LENGTH)


def generate_plain_variants():
    # test all lengths up to AES block_size + 1
    for i in range(Crypto.Cipher.AES.block_size + 2):
        yield crypto.random_bytes(i)


@pytest.mark.parametrize("plain", generate_plain_variants(), ids=len)
def test_encrypt_decrypt(AES_key, plain):
    # test normal flow
    encrypted = crypto.encrypt(plain, AES_key)
    decrypted = crypto.decrypt(encrypted, AES_key)
    assert decrypted == plain

    # test padding error
    with pytest.raises(crypto.IncorrectPadding):
        # decrement last byte in encrypted payload
        # to cause padding error while decrypting
        encrypted = bytearray(encrypted)

        # stay in byte-value range
        if encrypted[-1] > 0:
            encrypted[-1] -= 1
        else:
            encrypted[-1] = 0xFF

        crypto.decrypt(bytes(encrypted), AES_key)
