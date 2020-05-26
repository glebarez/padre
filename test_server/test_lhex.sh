cat testCiphers_lhex | go run ../... -u "http://localhost:5000/decrypt?enc=lhex" -p 200 -post "cipher=$" -e lhex "$@"
