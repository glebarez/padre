package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

var usage = `
	GoPaddy is a tool to exploit padding oracles, breaking CBC mode encryption.
	For details see link(https://en.wikipedia.org/wiki/Padding_oracle_attack)

Usage: cmd(GoPaddy [OPTIONS] [CIPHER])

CIPHER *required*
	the encoded (as plaintext) value of valid cipher, whose value is to be decrypted
	if not passed, GoPaddy will use STDIN, reading ciphers line by line
	The provided cipher will be internally decoded into bytes, 
	using specified encoder and replacement rules (see options: flag(-e), flag(-r))

OPTIONS:

flag(-u) *required*
	URL to request, use cipher($) character to define cipher placeholder for GET request.
	E.g. if URL is "http://vulnerable.com/?parameter=cipher($)"
	then HTTP request will be sent as "http://example.com/?parameter=cipher(payload)"
	the payload will be filled-in as a cipher, encoded using 
	specified encoder and replacement rules (see options: flag(-e), flag(-r))

flag(-err) *required*
	A padding error pattern, HTTP responses will be searched for this string to detect 
	padding oracle. Regex is supported (only response body is matched)

flag(-cookie)
	Cookie value to be set in HTTP reqeusts.
	Use cipher($) character to define cipher placeholder.

flag(-post)
	If you want GoPaddy to perform POST requests (instead of GET), 
	then provide string payload for POST request body in this parameter.
	Use cipher($) character to define cipher placeholder.
	The Content-Type will be determined automatically, based on provided data. 

flag(-ct)
	Content-Type header to be set in HTTP requests.
	If not specified, Content-Type will be determined automatically.
	Only applicable if POST requests are used (see flag(-post) options).
	
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
	contentType  *string
	cookies      []*http.Cookie
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
	_, err := fmt.Fprintln(color.Error, red(fmt.Sprintf("Parameter %s: %s", flag, text)))
	if err != nil {
		log.Fatal(err)
	}
	// set this flag
	hadErrors = true
}

func argWarning(flag string, text string) {
	var err error
	if flag != "" {
		_, err = fmt.Fprintln(color.Error, yellow(fmt.Sprintf("Parameter %s: %s", flag, text)))
	} else {
		_, err = fmt.Fprintln(color.Error, yellow(fmt.Sprintf("%s", text)))
	}
	if err != nil {
		log.Fatal(err)
	}
}

func parseCookies(cookies string) (cookSlice []*http.Cookie, err error) {
	// strip quotes if any
	cookies = strings.Trim(cookies, `"'`)

	// split several cookies into slice
	cookieS := strings.Split(cookies, ";")

	for _, c := range cookieS {
		// strip whitespace
		c = strings.TrimSpace(c)

		// split to name and value
		nameVal := strings.SplitN(c, "=", 2)
		if len(nameVal) != 2 {
			return nil, errors.New("failed to parse cookie")
		}

		cookSlice = append(cookSlice, &http.Cookie{Name: nameVal[0], Value: nameVal[1]})
	}
	return cookSlice, nil
}

func determineContentType(data string) string {
	var contentType string

	if data[0] == '{' || data[0] == '[' {
		contentType = "application/json"
	} else {
		match, _ := regexp.MatchString("([^=]*?=[^=]*?&?)+", data)
		if match {
			contentType = "application/x-www-form-urlencoded"
		} else {
			contentType = http.DetectContentType([]byte(data))
		}
	}
	return contentType
}

func parseArgs() (ok bool, cipher *string) {

	// set-up the flags
	config.URL = flag.String("u", "", "")
	encoding := flag.String("e", "b64", "")
	replacements := flag.String("r", "", "")
	cookies := flag.String("cookie", "", "")

	config.paddingError = flag.String("err", "", "")
	config.blockLen = flag.Int("b", 16, "")
	config.parallel = flag.Int("p", 100, "")
	config.proxyURL = flag.String("proxy", "", "")
	config.POSTdata = flag.String("post", "", "")
	config.contentType = flag.String("ct", "", "")

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

	// content-type
	if *config.POSTdata != "" && *config.contentType == "" {
		// if not passed, determine automatically
		var ct string
		ct = determineContentType(*config.POSTdata)
		config.contentType = &ct
		argWarning("", "Content-Type was determined automatically as: "+ct)
	}

	// cookies
	if *cookies != "" {
		config.cookies, err = parseCookies(*cookies)
		if err != nil {
			argError("-cookie", err.Error())
		}
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

	// decide on cipher source
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
		fmt.Fprintf(color.Error, fmt.Sprintf("\nRun with %s option to see usage help\n\n", cyanBold("-h")))
		ok = false
		return
	}

	ok = true
	return
}
