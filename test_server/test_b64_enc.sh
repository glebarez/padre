cat testPlains | go run ../... -u "http://localhost:5000/decrypt?cipher=$&enc=b64" -err IncorrectPadding -p 200 -post "cipher=$" -cookie "token=$" -enc "$@" > out
