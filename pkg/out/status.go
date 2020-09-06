package out

import (
	"errors"
	"os"
)

/* starts new hacky bar */
func (p *processingStatus) openBar(outputLen int) {
	// depending on mode (encryption or decryption), choose encoder for resulting byte data
	var encoder encoderDecoder
	if *config.encrypt {
		encoder = config.encoder
	} else {
		encoder = asciiEncoder{} // use ASCII decoder for decryption
	}

	// create bar
	p.bar = createHackyBar(&encoder, outputLen)

	// listen for events and reflect the status
	go p.bar.listenAndPrint()
}

/* gracefully shutting down the hackyBar goroutine */
func (p *processingStatus) closeBar() {
	p.bar.stop()

	// print warning if overflow occurred and stdout was not redirected
	if p.bar.overflow && isTerminal(os.Stdout) {
		printError(errors.New("Output was too wide to fit you terminal. Redirect stdout somewhere to get full output"))
	}
	p.bar = nil
}
