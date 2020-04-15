cat testPlains | go run ../... -u "http://localhost:5000/decrypt?cipher=$&enc=lhex" -err IncorrectPadding -p 200 -post "cipher=$" -cookie "token=$" -e lhex -enc "$@"
