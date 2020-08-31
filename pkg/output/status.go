package output

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

/* general status handle */
type processingStatus struct {
	prefix    string    // prefix of current status
	lineIndex int       // number of last line that was printed on
	fresh     bool      // indicator if status has already even printed something
	bar       *hackyBar // the dynamically changing, hacky-bar inside current status
}

/* global variables */
var (
	currentStatus   *processingStatus // currently active status
	totalInputs     int               // total count of inputs to be processed
	currentInputNum int               // id of currently processed input
)

/* creates new status handle */
func createNewStatus() {
	// increase the global status ID counter
	currentInputNum++

	// refresh the current global instance
	currentStatus = &processingStatus{
		fresh:  true,
		prefix: fmt.Sprintf("[%d/%d] ", currentInputNum, totalInputs),
	}
}

/* closes current status and prints newLine */
func closeCurrentStatus() {
	if currentStatus == nil {
		panic("No status currently open")
	}

	// print line-feed upon closing the status
	fmt.Fprintln(outputStream, "")
	currentStatus = nil
}

/* function to be used by external http client to report that yet-another request was made */
func reportHTTPRequest() {
	if currentStatus == nil {
		return
	}

	// http client can make request outside of bar scope (e.g. pre-flight checks), those do not count
	if currentStatus.bar != nil {
		currentStatus.bar.chanReq <- 1
	}
}

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

/* printing function in context of status */
func (p *processingStatus) print(s string, sameLine bool) {
	/* reset fresh flag after first print in current status
	the fresh flag is needed to avoid newLine printing when not necessary */
	defer func() {
		if p.fresh {
			p.fresh = false
		}
	}()

	// create builder for efficiency
	builder := &strings.Builder{}
	builder.Grow(config.termWidth)

	// if same-line print requested, clear the current contents of the line
	if sameLine {
		builder.WriteString("\x1b\x5b2K\r") // clear line + caret return
	} else {
		/* well, applying newLine logic to fresh instance is not necessary,
		otherwise we end up having a blank line */
		if !p.fresh {
			p.lineIndex++
			builder.WriteByte('\n')
		}
	}

	// add prefix only if it's the first line of current status
	if p.lineIndex == 0 {
		builder.WriteString(cyanBold(p.prefix))
	} else {
		// otherwise, just put spaces for same indent as prefix would do
		builder.WriteString(strings.Repeat(" ", len(p.prefix)))
	}

	// add the input string (the real payload)
	builder.WriteString(s)

	// output finally
	fmt.Fprint(outputStream, builder.String())
}

// fetches calculated output byte into hacky-bar
func (p *processingStatus) fetchOutputByte(b byte) {
	p.bar.chanOutput <- b
}

// fetches multiple calculated output bytes into hacky-bar
// uses reverse order
func (p *processingStatus) fetchOutputBytes(bs []byte) {
	for i := len(bs) - 1; i >= 0; i-- {
		p.fetchOutputByte(bs[i])
	}
}
