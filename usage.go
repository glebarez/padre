package main

import (
	"regexp"
)

var usage = `
Usage: cmd(padre [OPTIONS] [INPUT])

INPUT: 
	In bold(decrypt) mode: encrypted data
	In bold(encrypt) mode: the plaintext to be encrypted
	If not passed, will read from bold(STDIN)

	NOTE: binary data is always encoded in HTTP. Tweak encoding rules if needed (see options: flag(-e), flag(-r))

OPTIONS:

flag(-u) *required*
	target URL, use dollar($) character to define token placeholder (if present in URL)

flag(-enc)
	Encrypt mode

flag(-err)
	Regex pattern, HTTP response bodies will be matched against this to detect padding oracle. Omit to perform automatic fingerprinting

flag(-e)
	Encoding to apply to binary data. Supported values:
		b64 (standard base64) *default*
		lhex (lowercase hex)

flag(-r)
	Additional replacements to apply after encoding binary data. Use odd-length strings, consiting of pairs of characters <OLD><NEW>.
	Example:
		If server uses base64, but replaces '/' with '!', '+' with '-', '=' with '~', then use cmd(-r "/!+-=~")

flag(-cookie)
	Cookie value to be set in HTTP requests. Use dollar($) character to mark token placeholder.

flag(-post)
	String data to perform POST requests. Use dollar($) character to mark token placeholder. 

flag(-ct)
	Content-Type for POST requests. If not specified, Content-Type will be determined automatically.
	
flag(-b)
	Block length used in cipher (use 16 for AES). Supported values:
		8
		16 *default*
		32

flag(-p)
	Number of parallel HTTP connections established to target server [1-256]
		30 *default*
		
flag(-proxy)
	HTTP proxy. e.g. use cmd(-proxy "http://localhost:8080") for Burp or ZAP

bold(Examples:)
	Decrypt token in GET parameter:	cmd(padre -u "http://vulnerable.com/login?token=$" "u7bvLewln6PJ670Gnj3hnE40L0SqG8e6")
	POST data: cmd(padre -u "http://vulnerable.com/login" -post "token=$" "u7bvLewln6PJ670Gnj3hnE40L0SqG8e6")
	Cookies: cmd(padre -u "http://vulnerable.com/login$" -cookie "auth=$" "u7bvLewln6PJ670Gnj3hnE40L0SqG8e6")
	Encrypt token in GET parameter:	cmd(padre -u "http://vulnerable.com/login?token=$" -enc "EncryptMe")
`

func init() {
	// add some color to usage text
	re := regexp.MustCompile(`\*required\*`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(yellow(`(required)`))))

	re = regexp.MustCompile(`\*default\*`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(green(`(default)`))))

	re = regexp.MustCompile(`cmd\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(cyan("$1"))))

	re = regexp.MustCompile(`dollar\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(cyanBold("$1"))))

	re = regexp.MustCompile(`flag\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(greenBold("$1"))))

	re = regexp.MustCompile(`link\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(underline("$1"))))

	re = regexp.MustCompile(`bold\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(bold("$1"))))
}

// flag wrapper
func _f(f string) string {
	return `(` + greenBold(`-`+f) + ` option)`
}

var hint = struct {
	omitBlockLen     string
	omitErrPattern   string
	setErrPattern    string
	lowerConnections string
	checkEncoding    string
	checkInput       string
}{
	omitBlockLen:     `omit ` + _f(`b`) + `  for automatic detection of block length`,
	omitErrPattern:   `omit ` + _f(`err`) + ` for automatic fingerprinting of HTTP responses`,
	setErrPattern:    `specify error pattern manually with ` + _f(`err`),
	lowerConnections: `server might be overwhelmed or rate-limiting you requests. try lowering concurrency using ` + _f(`p`),
	checkEncoding:    `check that encoding ` + _f(`e`) + ` and replacement rules ` + _f(`r`) + ` are set properly`,
	checkInput:       `check that INPUT is properly formatted`,
}

func makeDetectionHints() []string {
	// hint intro
	intro := `if you believe target is vulnerable, try following:`
	li := cyanBold(`> `)
	hints := []string{intro}

	// block length
	if *config.blockLen != 0 {
		hints = append(hints, li+hint.omitBlockLen)
	} else {
		// error pattern
		if *config.paddingErrorPattern != "" {
			hints = append(hints, li+hint.omitErrPattern)
		} else {
			hints = append(hints, li+hint.setErrPattern)
		}
	}

	// concurrency
	if *config.parallel > 10 {
		hints = append(hints, li+hint.lowerConnections)
	}
	return hints
}
