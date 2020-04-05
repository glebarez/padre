# GoPaddy

A fast, Golang, concurrent tool to perform Padding Oracle attacks against CBC mode encryption.

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
All the fancy stuff (logo and progress tracking) is written to STDERR. <br>
Thus you may safely redirect STDOUT to a file, or pipe it with another tool. <br>
Only succesfully decrypted values will be written to redirection target.
You can supply (multiple) to-be-decrypted values into STDIN as well.

### Usage
```console
GoPaddy [OPTIONS] [CIPHER]
```

CIPHER:

	to-be-decrypted value (or token) that vulnerable server leaked to you.
	if not provided, values will be read from STDIN
	make sure you tip GoPaddy about the encoding nature with options -e and -r
						(e.g. base64-encoded ciphers)
	

OPTIONS:

-u

	Vilnerable URL (one that produces padding errors).
        Use $ character to define cipher placeholder for GET request.
	Example:
           if URL is "http://vulnerable.com/?parameter=$"
	   then HTTP request will be sent as "http://example.com/?parameter=payload"
	   the payload will be filled-in as a cipher, encoded using 
	   specified encoder and replacement rules (see options: -e, -r)

-err

	A padding error string as presented by the server.
        HTTP responses will be searched for this string to detect 
	padding oracle. Regex is supported.
        Only response body is matched.

-e

	Encoding, used by the server to encode binary ciphertext in HTTP context.
	This option is used in conjunction with -r option (see below)
	Supported values:
		b64 (standard base64) *default*
		lhex (lowercase hex)

-r

	Character replacement rules that server applies to encoded ciphertext.
	Use odd-length strings, consisting of character pairs: <OLD><NEW>.
	Example:
		If server uses base64, but replaces '/' with '!', '+' with '-', '=' with '~',
		then use -r "/!+-=~"

-cookie

	Cookie value to be decrypted in HTTP reqeusts.
	Use $ character to define cipher placeholder.

-post

	If you need to perform POST requests (instead of GET), 
	then provide string payload for POST request body in this parameter.
	Use $ character to define cipher placeholder.
        Example: 
                  -post "vuln_param=$"
          

-ct

	Content-Type to be set in HTTP requests.
        This option is only effective when -post option is used.
	If not specified, Content-Type will be inferred automatically
        based on data, provided in -post option.
	
-b

	Block length used in encrypting (always 16 for AES)
	Supported values:
		8u
		16 *default*
		32

-p

	Number of parallel HTTP connections established to target server [1-256]
		30 *default*
		
-proxy

	HTTP proxy. e.g. use -proxy "http://localhost:8080" for Burp or ZAP
