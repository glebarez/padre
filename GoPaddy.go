package main

import (
	"bufio"
	"os"

	"github.com/mattn/go-isatty"
)

func main() {

	printLogo()

	/* parse command line arguments, this will fill the config structure exit right away if not ok */
	ok, input := parseArgs()
	if !ok {
		return
	}

	// init HTTP client
	err := initHTTP()
	if err != nil {
		printError(err)
		os.Exit(1)
	}

	// build list of inputs
	inputs := make([]string, 0)

	// decide on whether read from STDIN
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

	initStatus(len(inputs))
	if *config.encrypt {
		// encrypt inputs one by one
		for _, c := range inputs {
			// create new status bar for every input
			createNewStatus()

			// encrypt
			cipher, err := encrypt(c)
			if err != nil {
				printError(err)
			}

			closeCurrentStatus()

			if err != nil {
				/* skip the rest for current input */
				continue
			}

			/* write output only if output is redirected to file
			because encrypted inputs already will be in status output
			and printing them again to STDOUT again, will be ugly */
			if !isTerminal(os.Stdout) {
				os.Stdout.Write(append([]byte(config.encoder.encode(cipher)), '\n'))
			}
		}
	} else {
		// decrypt inputs one by one
		for _, c := range inputs {
			// create new status bar
			createNewStatus()

			// decrypt
			plain, err := decrypt(c)
			if err != nil {
				printError(err)
			}

			closeCurrentStatus()

			if err != nil {
				/* skip the rest for current input */
				continue
			}

			/* write output only if output is redirected to file
			because decoded inputs already will be in status output
			and printing them again to STDOUT again, will be ugly */
			if !isTerminal(os.Stdout) {
				os.Stdout.Write(append(plain, '\n'))
			}
		}
	}

}

/* is terminal? */
func isTerminal(file *os.File) bool {
	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}
