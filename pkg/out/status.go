package out

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

	p.bar = nil
}
