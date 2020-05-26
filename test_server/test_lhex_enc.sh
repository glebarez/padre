cat testPlains | go run ../... -u "http://localhost:5000/decrypt?cipher=$&enc=lhex" -e lhex -enc "$@"
