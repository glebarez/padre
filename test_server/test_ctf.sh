cat testCiphersCTF.txt | go run ../... -u "http://34.74.105.127/499682b282/?post=$" -err PaddingException -r "/!+-=~" -p 20
