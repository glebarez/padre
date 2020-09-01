package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/glebarez/padre/pkg/encoder"
	"github.com/glebarez/padre/pkg/output"
	"github.com/glebarez/padre/pkg/util"
)

func init() {
	// a custom usage message
	flag.Usage = func() {
		output.Print(usage)
	}
}

const defaultConcurrency = 30

// Args - CLI flags
type Args struct {
	BlockLen            *int
	Parallel            *int
	TargetURL           *url.URL
	Encoder             encoder.Encoder
	PaddingErrorPattern *string
	ProxyURL            *url.URL
	POSTdata            *string
	ContentType         *string
	Cookies             []*http.Cookie
	TermWidth           int
	EncryptMode         *bool
	Input               *string
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

func parseArgs() (ok bool, args *Args) {
	var err error

	// create args struct
	args = &Args{}

	// simple flags that go in as-is
	args.PaddingErrorPattern = flag.String("err", "", "")
	args.BlockLen = flag.Int("b", 0, "")
	args.Parallel = flag.Int("p", defaultConcurrency, "")
	args.POSTdata = flag.String("post", "", "")
	args.ContentType = flag.String("ct", "", "")
	args.EncryptMode = flag.Bool("enc", false, "")

	// flags that need additional processing
	targetURL := flag.String("u", "", "")
	proxyURL := flag.String("proxy", "", "")
	encoding := flag.String("e", "b64", "")
	replacements := flag.String("r", "", "")
	cookies := flag.String("cookie", "", "")

	// parse flags
	flag.Parse()

	// general check on URL, POSTdata or Cookies for having the $ placeholder
	match1, err := regexp.MatchString(`\$`, *targetURL)
	if err != nil {
		argError("-u", err.Error())
	}
	match2, err := regexp.MatchString(`\$`, *args.POSTdata)
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

	// get terminal width
	args.TermWidth = util.TerminalWidth()
	if args.TermWidth == -1 {
		output.PrintWarning("Could not  determine your terminal width. Falling back to 80")
		args.TermWidth = 80 // fallback
	}

	// Target URL
	if *targetURL == "" {
		argError("-u", "Must be specified")
	} else {
		args.TargetURL, err = url.Parse(*targetURL)
		if err != nil {
			argError("-u", "Failed to parse URL: "+err.Error())
		}
	}

	// Proxy URL
	if *proxyURL != "" {
		args.ProxyURL, err = url.Parse(*proxyURL)
		if err != nil {
			argError("-proxy", "Failed to parse URL: "+err.Error())
		}
	}

	// Encoder (With replacements)
	if len(*replacements)%2 == 1 {
		argError("-r", "String must be of even length (0,2,4, etc.)")
	} else {
		switch strings.ToLower(*encoding) {
		case "b64":
			args.Encoder = encoder.NewB64encoder(*replacements)
		case "lhex":
			args.Encoder = encoder.NewLHEXencoder(*replacements)
		default:
			argError("-e", "Unsupported encoding specified")
		}
	}

	// block length
	switch *args.BlockLen {
	case 0: // = not set
	case 8:
	case 16:
	case 32:
	default:
		argError("-b", "Unsupported value passed. Omit, or specify one of: 8, 16, 32")
	}

	// Cookies
	if *cookies != "" {
		args.Cookies, err = util.ParseCookies(*cookies)
		if err != nil {
			argError("-cookie", fmt.Sprintf("Failed to parse cookies: %s", err))
		}
	}

	// Concurrency
	if *args.Parallel < 1 {
		argWarning("-p", fmt.Sprintf("Cannot be less than 1, value corrected to default value (%d)", defaultConcurrency))
		*args.Parallel = defaultConcurrency
	} else if *args.Parallel > 256 {
		argWarning("-p", "Cannot be greater than 256, value reduced to 256")
		*args.Parallel = 256
	}

	// decide on input source
	switch flag.NArg() {
	case 0:
		// no input passed, STDIN will be used
	case 1:
		// input is passed
		args.Input = &flag.Args()[0]
	default:
		// too many positional arguments
		argError("[INPUT]", "Specify exactly one input string, or pipe into STDIN")
	}

	// if errors in arguments, return here with message
	if hadErrors {
		output.Print(fmt.Sprintf("\nRun with %s option to see usage help\n", output.CyanBold("-h")))
		ok = false
		return
	}

	// show some info
	output.PrintInfo("padre is on duty")
	output.PrintInfo(fmt.Sprintf("Using concurrency (http connections): %d", *args.Parallel))

	// content-type detection
	if *args.POSTdata != "" && *args.ContentType == "" {
		// if not passed, determine automatically
		*args.ContentType = util.DetectContentType(*args.POSTdata)
		output.PrintSuccess("HTTP Content-Type detected automatically as " + output.Yellow(*args.ContentType))
	}

	ok = true
	return
}
