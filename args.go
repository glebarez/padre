package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"regexp"

	"github.com/fatih/color"
)

var usage = `
	GoPaddy is a fast tool to decrypt ciphers using padding oracle.
	For details: link(https://en.wikipedia.org/wiki/Padding_oracle_attack)

Usage: cmd(GoPaddy [OPTIONS] [CIPHER])

CIPHER *required*
	the encoded (to plaintext) value of valid cipher, whose value is to be decrypted
	if not passed, GoPaddy will use STDIN, reading ciphers line by line
	The provided cipher will be internally decoded into bytes, 
	using specified encoder (see option flag(-e))

OPTIONS:

flag(-u) *required*
	URL pattern to send, use cipher($) to define a cipher placeholder,
	e.g. if url is "http://vulnerable.com/?parameter=cipher($)"
	then HTTP request will be sent as "http://example.com/?parameter=cipher(payload)"
	the payload will be filled-in as a cipher, encoded using specified rules (see flag(-e) flag)

flag(-err) *required*
	A padding error pattern, HTTP responses will be searched for this string to detect 
	if padding exception has occured

flag(-b)
	Block length used in cipher (use 16 for AES)
	Supported values:
		8
		16 *default*
		32

flag(-e)
	Encoding/Decoding, used to translate encoded plaintext cipher into bytes (and back)
	When reading CIPHER, encoding is used backwards, to decode from plaintext to bytes
	Usually, cipher is encoded to enable passing as a plaintext URL parameter
	This option is used in conjunction with flag(-r) option (see below)
	Supported values:
		b64 (standard base64) *default*

flag(-r)
	Character replacement rules that vulnerable server applies
	after encoding ciphers to plaintext payloads.
	Use odd-length strings, consiting of pairs of characters <OLD><NEW>.
	Example:
		Generally, using standard base64 encoding is not suitable to pass ciphers
		in URL parameters. This is because standard base64 cotains characters: /,+,=
		Those have special meaning in URL syntax, therefore, some servers will
		further replace some of characters with others.
		E.g. if server uses base64, but replaces '/' with '!', '+' with '-', '=' with '~',
		then use cmd(-r "/!+-=~")
	link(NOTE:)
		these replacements will be internally applied in reverse direction
		before decoding plaintext cipher into bytes

flag(-p)
	Number of parallel HTTP connections established to target server
	The more connections, the faster is cracking speed
	If passed value is greater than 256, it will be reduced to 256
		100 *default*
		
flag(-proxy)
	HTTP proxy. e.g. use cmd(-proxy "http://localhost:8080") for Burp or ZAP
`

/* config structure is filled when command line arguments are parsed */
var config = struct {
	blockLen     *int
	parallel     *int
	URL          *string
	encoder      encoderDecoder
	paddingError *string
	proxyURL     *string
	POSTdata     *string
}{}

func init() {
	// add some color to usage text
	re := regexp.MustCompile(`\*required\*`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(yellow(`(required)`))))

	re = regexp.MustCompile(`\*default\*`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(greenBold(`(default)`))))

	re = regexp.MustCompile(`cmd\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(cyanBold("$1"))))

	re = regexp.MustCompile(`cipher\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(cyanBold("$1"))))

	re = regexp.MustCompile(`flag\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(greenBold("$1"))))

	re = regexp.MustCompile(`link\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(underline("$1"))))

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

func argWarning(flag string, text string) {
	_, err := color.New(color.FgYellow).Fprintf(color.Error, "Parameter %s: %s\n", flag, text)
	if err != nil {
		log.Fatal(err)
	}
}

func parseArgs() (ok bool, cipher *string) {

	// set-up the flags
	config.URL = flag.String("u", "", "")
	encoding := flag.String("e", "b64", "")
	replacements := flag.String("r", "", "")
	config.paddingError = flag.String("err", "", "")
	config.blockLen = flag.Int("b", 16, "")
	config.parallel = flag.Int("p", 100, "")
	config.proxyURL = flag.String("proxy", "", "")
	config.POSTdata = flag.String("post", "", "")

	// parse
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		return false, nil
	}

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
		argError("-e", "Unsupported encoding specified")
	}

	// padding error string
	if *config.paddingError == "" {
		argError("-err", "Must be specified")
	}

	// block length
	switch *config.blockLen {
	case 8:
	case 16:
	case 32:
	default:
		argError("-b", "Unsupported value passed")
	}

	// parallel connections
	if *config.parallel < 1 {
		argWarning("-p", "Cannot be less than 1, value corrected to default value (100)")
		*config.parallel = 100
	} else if *config.parallel > 256 {
		argWarning("-p", "Cannot be greater than 256, value corrected to 256")
		*config.parallel = 256
	}

	// general check on URL and POSTdata for having the $ placeholder
	match1, err := regexp.MatchString(`\$`, *config.URL)
	if err != nil {
		argError("-u", err.Error())
	}
	match2, err := regexp.MatchString(`\$`, *config.POSTdata)
	if err != nil {
		argError("-post", err.Error())
	}
	if !(match1 || match2) {
		argError("-u, -post", "Either URL or POST data must contain the $ placeholder")
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
		argError("[CIPHER]", "Too many arguments specified, specifiy exactly one string, or leave empty to read from STDIN")
	}

	// print info about errors occured
	if hadErrors {
		fmt.Fprintf(color.Error, fmt.Sprintf("run %s to see usage help\n", greenBold("GoPaddy -h")))
		ok = false
		return
	}

	ok = true
	return
}
