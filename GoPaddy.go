package main

import (
	"bufio"
	"os"

	"github.com/mattn/go-isatty"
)

var config = struct {
	blockLen     *int
	parallel     *int
	URL          *string
	encoder      encoderDecoder
	paddingError *string
	proxyURL     *string
}{}

// var baseURL = "http://localhost:5000/decrypt?cipher=%s"
// var cipherEncoded = "jigNcuWcyzd8QB7E/fm7peYSX9gnh6/gYG5Hmy/Bz7IVHVUM1hFyoCjPREV5efzK"
// var paddingError = "IncorrectPadding"

// var baseURL = "http://35.227.24.107/7631b88aa5/?post=%s"
// var cipherEncoded = "SqSdDHQt0u3b3Hmzklmd2oom2AjfJ8gmwir8PPXBXy6ybHE1o3KRleVxELoZAu-7MiAJGNCV075GhBsdokAFm0JLMA9XHJ4SLCIRU7K!6HktXt!y9rD4MEf6kvzxftlt35jGUuqL3t0RwSJjcMC-7eQuN9aFue5p9kqA7MlQSUiSD0J9Id8mCqsbwLXGohGS5w53EJz9jX6-g1vkS3lDiA~~"
// var paddingError = "PaddingException"

func main() {
	printLogo()

	// parse command line arguemnts, this will fill the config strucutre described above
	// exit if not ok
	ok, cipher := parseArgs()
	if !ok {
		return
	}

	// init HTTP client
	initHTTP()

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
	for _, c := range ciphers {
		createStatus()
		plain, err := decipher(c)
		if err != nil {
			currentStatus.error(err)
		}

		// write output only of output is redirected to file
		// beacuse decoded ciphers already will be in status output
		// and printing them once again, will be ugly
		if !isatty.IsTerminal(os.Stdout.Fd()) {
			os.Stdout.Write(plain)
		}
	}

}
