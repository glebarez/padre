package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/fatih/color"
)

var usage = `
	GoPaddy is a fast tool to decrypt ciphers using padding oracle.
	For details: https://en.wikipedia.org/wiki/Padding_oracle_attack

Usage: ` + greenBold("gopaddy OPTIONS CIPHER") + `

CIPHER (REQUIRED):
	the encoded (to plaintext) value of valid cipher, whose value is to be decrypted
	if not passed, GoPaddy will use STDIN, reading ciphers line by line,
	which allows deciphering of multiple inputs in one run.
	The provided cipher(s) will be internally decoded into bytes, 
	using specified encoder (see option -e)

OPTIONS:
-u (REQUIRED)
	URL pattern to send, use "` + cyanBold(`$`) + `" to define a cipher placeholder,
	e.g. if url is "http://vulnerable.com/?parameter=` + cyanBold("$") + `"
	then HTTP request will be sent as "http://example.com/?parameter=` + cyanBold(`payload`) + `"
	the payload will be filled-in as a cipher, encoded using specified rules (see -e flag)
-err (REQUIRED)
	A padding error pattern, HTTP responses will be searched for this string to detect 
	if padding exception has occured
-b
	Block length used in cipher (use 16 for AES)
	Supported values:
		8
		16 (DEFAULT)
		32
-enc
	Encoding/Decoding, used to translate encoded plaintext cipher into bytes (and back)
	When reading CIPHER, encoding is used backwards, to decode from plaintext to bytes
	Usually, cipher is encoded to enable passing as a plaintext URL parameter
	This option is used in conjunction with -r option (see below)
	Supported values:
		b64 (standard base64) (DEFAULT)
		hex
-r
	Character replacement rules that vulnerable server applies
	after encoding ciphers to plaintext payloads.
	Use odd-length strings, consiting of pairs of characters <OLD><NEW>.
	Example:
		Generally, using standard base64 encoding is not suitable to pass ciphers
		in URL parameters. This is because standard base64 cotains characters: /,+,=
		Those have special meaning in URL syntax, therefore, some servers will
		further replace some of characters with others.
		E.g. if server uses base64, but replaces '/' with '!', '+' with '-', '=' with '~',
		then use -r "/!+-=~"
	NOTE:
		these replacements will be internally applied in reverse direction
		before decoding plaintext cipher into bytes
-p
	Number of parallel HTTP connections established to target server
	The more connections, the faster is cracking speed
	If passed value is greater than 256, it will be reduced to 256
		DEFAULT: 50
-proxy
	HTTP proxy. e.g. use "http://localhost:8080" for Burp or ZAP
`

func init() {
	// a custom usage
	flag.Usage = func() {
		fmt.Fprintf(color.Error, usage)
	}
}

// error printing function, used when checking passed values
var hadErrors bool

func argError(flag string, text string) {
	_, err := color.New(color.FgRed, color.Bold).Fprintf(color.Error, "Parameter %s: %s\n", flag, text)
	if err != nil {
		log.Fatal(err)
	}
	// set this flag
	hadErrors = true
}

func parseArgs() (ok bool, cipher *string) {

	// set-up the flags
	config.URL = flag.String("u", "", "")
	encoding := flag.String("enc", "b64", "")
	replacements := flag.String("r", "", "")
	config.blockLen = flag.Int("b", 16, "")
	config.parallel = flag.Int("p", 50, "")
	config.paddingError = flag.String("err", "", "")
	config.proxyURL = flag.String("proxy", "", "")

	// parse
	flag.Parse()

	// check values
	var err error

	// URL
	if *config.URL == "" {
		argError("-u", "Must be specified")
	} else {
		_, err = url.Parse(*config.URL)
		if err != nil {
			argError("-u", "Failed to parse URL: "+err.Error())
		}
	}

	// encoding and replacement
	switch *encoding {
	case "b64":
		config.encoder, err = createBase64EncoderDecoder(*replacements)
		if err != nil {
			argError("-r", err.Error())
		}
	default:
		argError("-enc", "Unsupported value passed")
	}

	// padding error string
	if *config.paddingError == "" {
		argError("-err", "Must be specified")
	}

	// decide on cipher used
	switch flag.NArg() {
	case 0:
		// no cipher passed, STDIN will be used
	case 1:
		// cipher is passed
		cipher = &flag.Args()[0]
	default:
		// too many positional arguments
		argError("CIPHER", "Too many arguments specified, specifiy exactly one string, or leave empty to read from STDIN")
	}

	if hadErrors {
		fmt.Fprintf(color.Error, fmt.Sprintf("run %s to see usage help", greenBold("gopaddy -h")))
		ok = false
		return
	}
	ok = true
	return
}
