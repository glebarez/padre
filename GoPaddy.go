package main

import (
	"bufio"
	"log"
	"os"

	"net/http"
	_ "net/http/pprof"

	"github.com/mattn/go-isatty"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// time.Sleep(10 * time.Second)
	// defer time.Sleep(10 * time.Second)

	printLogo()

	/* parse command line arguemnts, this will fill the config structure
	exit right away if not ok */
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
	initStatus(len(ciphers))
	for _, c := range ciphers {
		// create new status bar
		createNewStatus()

		// decipher
		plain, err := decipher(c)
		if err != nil {
			printError(err)
			continue
		}

		/* write output only if output is redirected to file
		because decoded ciphers already will be in status output
		and printing them again to STDOUT again, will be ugly */
		if !isTerminal(os.Stdout) {
			os.Stdout.Write(append(plain, '\n'))
		}
	}

	// flush output afterwards, just in case
	if !isTerminal(os.Stdout) {
		err = os.Stdout.Sync()
		if err != nil {
			printError(err)
			os.Exit(1)
		}
	}
}

/* is terminal? */
func isTerminal(file *os.File) bool {
	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}
