package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/glebarez/padre/pkg/color"
	"github.com/glebarez/padre/pkg/encoder"
	"github.com/glebarez/padre/pkg/util"
)

func init() {
	// a custom usage message
	flag.Usage = func() {
		fmt.Fprint(stderr, usage)
	}
}

const (
	defaultConcurrency   = 30
	defaultTerminalWidth = 80
	maxConcurrency       = 256
)

// Args - CLI flags
type Args struct {
	BlockLen            *int
	Parallel            *int
	TargetURL           *string
	Encoder             encoder.Encoder
	PaddingErrorPattern *string
	ProxyURL            *url.URL
	POSTdata            *string
	ContentType         *string
	Cookies             []*http.Cookie
	EncryptMode         *bool
	Input               *string
}

func parseArgs() (*Args, *argErrors) {
	// container for storing errors and warnings
	argErrs := newArgErrors()

	args := &Args{}

	// simple flags that go in as-is
	args.PaddingErrorPattern = flag.String("err", "", "")
	args.BlockLen = flag.Int("b", 0, "")
	args.Parallel = flag.Int("p", defaultConcurrency, "")
	args.POSTdata = flag.String("post", "", "")
	args.ContentType = flag.String("ct", "", "")
	args.EncryptMode = flag.Bool("enc", false, "")
	args.TargetURL = flag.String("u", "", "")

	// flags that need additional processing
	proxyURL := flag.String("proxy", "", "")
	encoding := flag.String("e", "b64", "")
	replacements := flag.String("r", "", "")
	cookies := flag.String("cookie", "", "")

	// parse flags
	flag.Parse()

	// general check on URL, POSTdata or Cookies for having the $ placeholder
	match1, err := regexp.MatchString(`\$`, *args.TargetURL)
	if err != nil {
		argErrs.flagError("-u", err)
	}
	match2, err := regexp.MatchString(`\$`, *args.POSTdata)
	if err != nil {
		argErrs.flagError("-post", err)
	}
	match3, err := regexp.MatchString(`\$`, *cookies)
	if err != nil {
		argErrs.flagError("-cookie", err)
	}
	if !(match1 || match2 || match3) {
		argErrs.flagErrorf("-u, -post, -cookie", "Either URL, POST data or Cookie must contain the $ placeholder")
	}

	// Target URL
	if *args.TargetURL == "" {
		argErrs.flagErrorf("-u", "Must be specified")
	} else {
		_, err = url.Parse(*args.TargetURL)
		if err != nil {
			argErrs.flagError("-u", fmt.Errorf("failed to parse URL: %w", err))
		}
	}

	// Proxy URL
	if *proxyURL != "" {
		args.ProxyURL, err = url.Parse(*proxyURL)
		if err != nil {
			argErrs.flagError("-proxy", fmt.Errorf("failed to parse URL: %w", err))
		}
	}

	// Encoder (With replacements)
	if len(*replacements)%2 == 1 {
		argErrs.flagErrorf("-r", "String must be of even length (0,2,4, etc.)")
	} else {
		switch strings.ToLower(*encoding) {
		case "b64":
			args.Encoder = encoder.NewB64encoder(*replacements)
		case "lhex":
			args.Encoder = encoder.NewLHEXencoder(*replacements)
		default:
			argErrs.flagErrorf("-e", "Unsupported encoding specified")
		}
	}

	// block length
	switch *args.BlockLen {
	case 0: // = not set
	case 8:
	case 16:
	case 32:
	default:
		argErrs.flagErrorf("-b", "Unsupported value passed. Omit, or specify one of: 8, 16, 32")
	}

	// Cookies
	if *cookies != "" {
		args.Cookies, err = util.ParseCookies(*cookies)
		if err != nil {
			argErrs.flagError("-cookie", fmt.Errorf("failed to parse cookies: %s", err))
		}
	}

	// Concurrency
	if *args.Parallel < 1 {
		argErrs.flagWarningf("-p", "Cannot be less than 1, value corrected to default value (%d)", defaultConcurrency)
		*args.Parallel = defaultConcurrency
	} else if *args.Parallel > maxConcurrency {
		argErrs.flagWarningf("-p", "Value reduced to maximum allowed value (%d)", maxConcurrency)
		*args.Parallel = maxConcurrency
	}

	// content-type auto-detection
	if *args.POSTdata != "" && *args.ContentType == "" {
		*args.ContentType = util.DetectContentType(*args.POSTdata)
		argErrs.warningf("HTTP Content-Type detected automatically as %s", color.Yellow(*args.ContentType))
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
		argErrs.flagErrorf("[INPUT]", "Specify exactly one input string, or pipe into STDIN")
	}

	return args, argErrs
}
