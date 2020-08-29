from argparse import Namespace

import pytest

from app import app
from encoder import Encoding


@pytest.fixture
def client(is_vulnerable: bool):
    # create app config
    config = Namespace(VULNERABLE=is_vulnerable)

    # inject config
    app.config.from_object(config)

    # create test client
    return app.test_client()


@pytest.fixture
def call_route(client, http_method):
    if http_method == "GET":
        return client.get
    elif http_method == "POST":
        return client.post
    else:
        raise AssertionError("Not supported HTTP method: %s" % http_method)


@pytest.mark.parametrize("is_vulnerable", [True, False])
@pytest.mark.parametrize("http_method", ["GET", "POST"])
@pytest.mark.parametrize("encoding", list(Encoding))
@pytest.mark.parametrize("plaintext", [""])
def test_app(call_route, plaintext, is_vulnerable, encoding):
    # send plaintext for encryption
    resp = call_route("/encrypt", data={"plain": plaintext, "enc": encoding.name})
    assert resp.status_code == 200

    # get response string
    cipher = resp.data.decode()

    # send for decryption
    resp = call_route("/decrypt", data={"cipher": cipher, "enc": encoding.name})
    assert resp.status_code == 200

    # compare results
    deciphered = resp.data.decode()
    assert deciphered == plaintext

    # send malformed cipher
    malformed_cipher = cipher[:-1]
    resp = call_route("/decrypt", data={"cipher": malformed_cipher})
    assert resp.status_code == 500

    # check response verbosity
    if not is_vulnerable:
        assert resp.data.decode() == "Internal server error"
    else:
        assert "Traceback" in resp.data.decode()


@pytest.mark.parametrize("is_vulnerable", [True, False])
@pytest.mark.parametrize("http_method", ["GET", "POST"])
def test_absent_params(call_route):
    # no plaintext
    resp = call_route("/encrypt")
    assert resp.status_code == 500

    # no ciphertext
    resp = call_route("/decrypt")
    assert resp.status_code == 500

    # no explicit encoding
    resp = call_route("/encrypt", data={"plain": "test"})
    assert resp.status_code == 200
