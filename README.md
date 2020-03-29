# GoPaddy

A golang, concurrent tool to perform Padding Oracle attacks against CBC mode encryption.

![demo](demo.gif)


### Installation
```console
go get github.com/glebarez/GoPaddy
```

### Examples
- Decipher tokens in GET parameters
```console
GoPaddy -u "http://vulnerable.com/login?token=$" -err "Invalid padding" "u7bvLewln6PJ670Gnj3hnE40L0SqG8e6"
````

- POST data
```console
GoPaddy -u "http://vulnerable.com/login" -post "token=$" -err "Invalid padding" "u7bvLewln6PJ670Gnj3hnE40L0SqG8e6"
````

- Cookies
```console
GoPaddy -u "http://vulnerable.com/login" -cookie "auth=$" -err "Invalid padding" "u7bvLewln6PJ670Gnj3hnE40L0SqG8e6"
````
### Note on tool chaining
All the fancy stuff (logo and progress tracking) is spilled to STDERR. <br>
You may safely redirect tool's STDOUT to a file or pipe it with another tool. <br>
Only succesfully decrypted values will be written to redirection target.

### Usage
```console
GoPaddy [OPTIONS] [CIPHER]
```

CIPHER:

	to-be-decrypted value (or token) that a server leaked to you.
	if not provided, values will be read from STDIN
	make sure you tip GoPaddy about the encoding nature with options -e and -r
						(e.g. base64-encoded ciphers)
	

OPTIONS:

-u

	URL to request, use $ character to define cipher placeholder for GET request.
	E.g. if URL is "http://vulnerable.com/?parameter=$"
	then HTTP request will be sent as "http://example.com/?parameter=payload"
	the payload will be filled-in as a cipher, encoded using 
	specified encoder and replacement rules (see options: -e, -r)

-err

	A padding error pattern, HTTP responses will be searched for this string to detect 
	padding oracle. Regex is supported (only response body is matched)

-e

	Encoding that server uses to present cipher as plaintext in HTTP context.
	This option is used in conjunction with -r option (see below)
	Supported values:
		b64 (standard base64) *default*

-r

	Character replacement rules that vulnerable server applies
	after encoding ciphers to plaintext payloads.
	Use odd-length strings, consiting of pairs of characters <OLD><NEW>.
	Example:
		If server uses base64, but replaces '/' with '!', '+' with '-', '=' with '~',
		then use -r "/!+-=~"

-cookie

	Cookie value to be set in HTTP reqeusts.
	Use $ character to define cipher placeholder.

-post

	If you want GoPaddy to perform POST requests (instead of GET), 
	then provide string payload for POST request body in this parameter.
	Use $ character to define cipher placeholder. 

-ct

	Content-Type header to be set in HTTP requests.
	If not specified, Content-Type will be determined automatically.
	Only applicable if POST requests are used (see -post options).
	
-b

	Block length used in cipher (use 16 for AES)
	Supported values:
		8
		16 *default*
		32

-p

	Number of parallel HTTP connections established to target server [1-256]
		30 *default*
		
-proxy

	HTTP proxy. e.g. use -proxy "http://localhost:8080" for Burp or ZAP
