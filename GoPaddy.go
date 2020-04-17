package main

import (
	"bufio"
	"os"
)

func main() {
	/* parse command line arguments, this will fill the config structure exit right away if not ok */
	ok, input := parseArgs()
	if !ok {
		return
	}

	/* of course some branding */
	if !*config.nologo {
		printLogo()
	}

	/* initialize HTTP client */
	err := initHTTP()
	if err != nil {
		printError(err)
		os.Exit(1)
	}

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

	/* globally initialize status subsystem */
	initStatus(len(inputs))

	/* general processor function to avoid copy-paste later */
	var do func(string) ([]byte, error)
	if *config.encrypt {
		do = encrypt
	} else {
		do = decrypt
	}

	/* process inputs one by one */
	var errCount int // error counter

	for _, c := range inputs {
		// create new status bar for every input
		createNewStatus()

		// do the work
		output, err := do(c)
		if err != nil {
			printError(err)
			errCount++
		}

		// close status for current input
		closeCurrentStatus()

		/* in case of error, skip to the next input */
		if err != nil {
			continue
		}

		/* write output only if output is redirected to file
		because encrypted inputs already will be in status output
		and printing them again to STDOUT again, will be ugly */
		if !isTerminal(os.Stdout) {
			/* in case of encryption, additionally encode the produced output */
			if *config.encrypt {
				output = []byte(config.encoder.encode(output))
			}

			/* write to standard output */
			os.Stdout.Write(append(output, '\n'))
		}
	}

	/* non-zero return code if all inputs were errornous */
	if len(inputs) == errCount {
		os.Exit(2)
	}

}
