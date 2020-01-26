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
GoPaddy -u "http://vulnerable.com/login$" -cookie "auth=$" -err "Invalid padding" "u7bvLewln6PJ670Gnj3hnE40L0SqG8e6"
````

### Usage
```console
GoPaddy [OPTIONS] [CIPHER]
```

CIPHER:

	the encoded (as plaintext) value of valid cipher, whose value is to be decrypted
	if not passed, GoPaddy will use STDIN, reading ciphers line by line
	The provided cipher must be encoded as specified in flag(-e) and flag(-r) options. 

OPTIONS:

-u

	URL to request, use cipher($) character to define cipher placeholder for GET request.
	E.g. if URL is "http://vulnerable.com/?parameter=cipher($)"
	then HTTP request will be sent as "http://example.com/?parameter=cipher(payload)"
	the payload will be filled-in as a cipher, encoded using 
	specified encoder and replacement rules (see options: flag(-e), flag(-r))

-err

	A padding error pattern, HTTP responses will be searched for this string to detect 
	padding oracle. Regex is supported (only response body is matched)

-e

	Encoding that server uses to present cipher as plaintext in HTTP context.
	This option is used in conjunction with flag(-r) option (see below)
	Supported values:
		b64 (standard base64) *default*

-r

	Character replacement rules that vulnerable server applies
	after encoding ciphers to plaintext payloads.
	Use odd-length strings, consiting of pairs of characters <OLD><NEW>.
	Example:
		If server uses base64, but replaces '/' with '!', '+' with '-', '=' with '~',
		then use cmd(-r "/!+-=~")

-cookie

	Cookie value to be set in HTTP reqeusts.
	Use cipher($) character to define cipher placeholder.

-post

	If you want GoPaddy to perform POST requests (instead of GET), 
	then provide string payload for POST request body in this parameter.
	Use cipher($) character to define cipher placeholder. 

-ct

	Content-Type header to be set in HTTP requests.
	If not specified, Content-Type will be determined automatically.
	Only applicable if POST requests are used (see flag(-post) options).
	
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

	HTTP proxy. e.g. use cmd(-proxy "http://localhost:8080") for Burp or ZAP