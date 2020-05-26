cat testPlains | go run ../... -u "http://localhost:5000/decrypt?cipher=$&enc=b64" -p 200 -enc "$@"
