package main

import (
	"bufio"
	"os"

	"github.com/mattn/go-isatty"
)

/* config structure is filled when command line arguments are parsed */
var config = struct {
	blockLen     *int
	parallel     *int
	URL          *string
	encoder      encoderDecoder
	paddingError *string
	proxyURL     *string
}{}

func main() {
	printLogo()

	/* parse command line arguemnts, this will fill the config strucutre described above
	exit if not ok */
	ok, cipher := parseArgs()
	if !ok {
		return
	}

	// init HTTP client
	err := initHTTP()
	if err != nil {
		printError(err)
		os.Exit(1)
	}

	// build list of ciphers
	ciphers := make([]string, 0)

	// decide on whether read from STDIN
	if cipher == nil {
		// read ciphers from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			ciphers = append(ciphers, scanner.Text())
		}
	} else {
		// use single cipher, passed in command line
		ciphers = append(ciphers, *cipher)
	}

	// crack ciphers one by one
	for i, c := range ciphers {
		// create new status bar with prefix of cipher order number
		createStatus(i+1, len(ciphers))

		// decipher
		plain, err := decipher(c)
		if err != nil {
			currentStatus.error(err)
			continue
		}

		/* write output only if output is redirected to file
		because decoded ciphers already will be in status output
		and printing them again to STDOUT again, will be ugly */
		if !(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())) {
			os.Stdout.Write(append(plain, '\n'))
		}
	}

	// flush output afterwards
	if !(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())) {
		err = os.Stdout.Sync()
		if err != nil {
			printError(err)
			os.Exit(1)
		}
	}
}
