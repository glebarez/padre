package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/glebarez/padre/pkg/color"
	"github.com/glebarez/padre/pkg/util"
	"honnef.co/go/tools/config"
)

func main() {
	// parse CLI arguments, exit if not ok
	ok, args := parseArgs()
	if !ok {
		os.Exit(1)
	}

	// show welcomming message
	out.PrintInfo("padre is on duty")
	out.PrintInfo(fmt.Sprintf("Using concurrency (http connections): %d", *args.Parallel))

	// content-type detection
	if *args.POSTdata != "" && *args.ContentType == "" {
		// if not passed, determine automatically
		*args.ContentType = util.DetectContentType(*args.POSTdata)
		out.PrintSuccess("HTTP Content-Type detected automatically as " + out.Yellow(*args.ContentType))
	}

	ok = true
	return

	var err error

	/* parse command line arguments, this will fill the config structure exit right away if not ok */
	ok, input := parseArgs()
	if !ok {
		return
	}

	/* initialize HTTP client */
	err = initHTTP()
	if err != nil {
		die(err)
	}

	// detect/confirm padding oracle
	if *config.blockLen != 0 {
		if err = detectOrConfirmPaddingOracle(*config.blockLen); err != nil {
			err := newErrWithHints(err, makeDetectionHints()...)
			die(err)
		}
	} else {
		// if block length is not set explicitly, detect it by testing possible values
		for _, blockLen := range []int{8, 16, 32} {
			if err = detectOrConfirmPaddingOracle(blockLen); err == nil {
				printSuccess(fmt.Sprintf("Detected block length: %d", blockLen))
				*config.blockLen = blockLen
				break
			}
		}
		if err != nil {
			err := newErrWithHints(err, makeDetectionHints()...)
			die(err)
		}
	}

	/* choose processing  processing function depending on the mode*/
	var do func(string) ([]byte, error)
	var mode string

	if *config.encrypt {
		do = encrypt
		mode = "encrypt"
	} else {
		do = decrypt
		mode = "decrypt"
	}
	printWarning("Mode: " + cyanBold(mode))

	/* process inputs one by one */
	var errCount int

	/* build list of inputs */
	inputs := make([]string, 0)

	if input == nil {
		// read inputs from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputs = append(inputs, scanner.Text())
		}
	} else {
		// use single input, passed in command line
		inputs = append(inputs, *input)
	}

	totalInputs = len(inputs)

	for _, c := range inputs {
		// create new status bar for current input
		createNewStatus()

		// do the work
		output, err := do(c)

		// in case of error, skip to the next input
		if err != nil {
			printError(err)
			errCount++
		}

		// close status for current input
		closeCurrentStatus()

		// continue to the next input
		if err != nil {
			continue
		}

		/* write output only if output is redirected to file
		because outputs already will be in status output
		and printing them again to STDOUT again, will be ugly */
		if !isTerminal(os.Stdout) {
			var err error
			/* in case of encryption, additionally encode the produced output */
			if *config.encrypt {
				outputStr := config.encoder.EncodeToString(output)
				_, err = os.Stdout.WriteString(outputStr + "\n")
			} else {
				os.Stdout.Write(append(output, '\n'))
			}

			if err != nil {
				die(err)
			}
		}
	}

	/* non-zero return code if all inputs were errornous */
	if len(inputs) == errCount {
		os.Exit(2)
	}
}

// detects padding error fingerprint if matching pattern is not provided
// otherwise, confirms padding oracle using provided pattern
func detectOrConfirmPaddingOracle(blockLen int) error {
	if *config.paddingErrorPattern == "" {
		printAction("Fingerprinting HTTP responses for padding oracle...")
		fp, err := detectPaddingErrorFingerprint(blockLen)
		if err != nil {
			return err
		}
		if fp != nil {
			printSuccess("Successfully detected padding oracle")
			config.paddingErrorFingerprint = fp
		} else {
			return fmt.Errorf("failed to auto-detect padding oracle response")
		}
	} else {
		printAction("Confirming padding oracle...")
		confirmed, err := confirmPaddingOracle(blockLen)
		if err != nil {
			return err
		}
		if confirmed {
			printSuccess("Confirmed padding oracle")
		} else {
			return fmt.Errorf("padding oracle not confirmed")
		}
	}
	return nil
}

func makeDetectionHints(*config.Config) []string {
	// hint intro
	intro := `if you believe target is vulnerable, try following:`
	li := color.CyanBold(`> `)
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
