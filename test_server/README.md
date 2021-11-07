## Test server that is (on-demand) vulnerable to Padding Oracle
Use for testing purposes. AES only by now

## Config
Configuration is done via setting environment variables
|Env. variable | if not set | if set |
|---|---|---|
|VULNERABLE|Server **is not** vulnerable to padding oracle|Server **is** vulnerable to padding oracle|
|SECRET|AES key will be generated randomly|AES key will generated from the secret phrase. Use to achieve reproducible outputs between server runs|
|USE_GEVENT|Use Flask's built-in Web server|Use gevent's Web server (faster)

## Run
#### via Docker Compose
```console
docker-compose up
```
#### via Docker
```console
docker build -t pador_vuln_server .
docker run -it -p 5000:5000 pador_vuln_server
```
#### directly
```console
python server.py
```



