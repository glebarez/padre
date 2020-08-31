package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/glebarez/padre/pkg/encoding"
	phttp "github.com/glebarez/padre/pkg/http"
	"github.com/glebarez/padre/pkg/output"
)

func init() {
	// a custom usage
	flag.Usage = func() {
		output.Print(usage)
	}
}

const defaultConcurrency = 30

// Args - CLI flags
type Args struct {
	BlockLen                *int
	Parallel                *int
	URL                     *string
	Encoding                encoding.EncoderDecoder
	PaddingErrorPattern     *string
	PaddingErrorFingerprint *phttp.ResponseFingerprint
	ProxyURL                *string
	POSTdata                *string
	ContentType             *string
	Cookies                 []*http.Cookie
	TermWidth               int
	EncryptMode             *bool
}

/* overall indicator for flag-parsing errors */
var hadErrors bool

func argError(flag string, text string) {
	output.PrintError(fmt.Errorf("Parameter %s: %s", flag, text))
	hadErrors = true
}

func argWarning(flag string, text string) {
	output.PrintWarning(fmt.Sprintf("Parameter %s: %s", flag, text))
}

func parseArgs(ok bool, input *string) *config.Config {
	// create args struct
	args := Args{}

	// setup flags
	args.URL = flag.String("u", "", "")
	args.PaddingErrorPattern = flag.String("err", "", "")
	args.BlockLen = flag.Int("b", 0, "")
	args.Parallel = flag.Int("p", defaultConcurrency, "")
	args.ProxyURL = flag.String("proxy", "", "")
	args.POSTdata = flag.String("post", "", "")
	args.ContentType = flag.String("ct", "", "")
	args.EncryptMode = flag.Bool("enc", false, "")

	// not-yet config flags
	encoding := flag.String("e", "b64", "")
	replacements := flag.String("r", "", "")
	cookies := flag.String("cookie", "", "")

	// parse flags
	flag.Parse()

	// get terminal width
	config.termWidth = terminalWidth()
	if config.termWidth == -1 {
		printWarning("Could not  determine your terminal width. Falling back to 80")
		config.termWidth = 80 // fallback
	}

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

	// encoding and replacement rules
	if len(*replacements)%2 == 1 {
		argError("-r", "String must be of even length")
	} else {
		switch strings.ToLower(*encoding) {
		case "b64":
			config.encoder = createB64encDec(*replacements)
		case "lhex":
			config.encoder = createLHEXencDec(*replacements)
		default:
			argError("-e", "Unsupported encoding specified")
		}
	}

	// block length
	switch *config.blockLen {
	case 0: // = not set
	case 8:
	case 16:
	case 32:
	default:
		argError("-b", "Unsupported value passed. Specify one of: 8, 16, 32")
	}

	// parallel connections
	if *config.parallel < 1 {
		argWarning("-p", fmt.Sprintf("Cannot be less than 1, value corrected to default value (%d)", defaultConcurrency))
		*config.parallel = defaultConcurrency
	} else if *config.parallel > 256 {
		argWarning("-p", "Cannot be greater than 256, value corrected to 256")
		*config.parallel = 256
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

	// decide on input source
	switch flag.NArg() {
	case 0:
		// no input passed, STDIN will be used
	case 1:
		// input is passed
		input = &flag.Args()[0]
	default:
		// too many positional arguments
		argError("[INPUT]", "Specify exactly one input string, or pipe into STDIN")
	}

	// if errors in arguments, return here with message
	if hadErrors {
		fmt.Fprintf(outputStream, fmt.Sprintf("\nRun with %s option to see usage help\n\n", cyanBold("-h")))
		ok = false
		return
	}

	// show some info
	printInfo("padre is on duty")
	printInfo(fmt.Sprintf("Using concurrency (http connections): %d", *config.parallel))

	// content-type detection
	if *config.POSTdata != "" && *config.contentType == "" {
		// if not passed, determine automatically
		var ct string
		ct = detectContentType(*config.POSTdata)
		config.contentType = &ct
		printSuccess("HTTP Content-Type detected automatically: " + yellow(ct))
	}

	ok = true
	return
}
