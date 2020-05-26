cat testCiphers | go run ../... -u "http://localhost:5000/decrypt" -p 200 -post "cipher=$" "$@"
