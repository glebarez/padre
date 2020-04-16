package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const (
	updateFreq     = 13   /* status update frequency (times/sec) */
	ccREPLACEBYTES = 9000 /* output bytes-stream control magic number */
)

/* global variables */
var (
	currentStatus *processingStatus // currently active status (we use singleton for now)
	totalCount    int               // total count of all statuses to be shown
	currentID     int               // id of currently active status
)

/* global init, to tell how many inputs/statuses is there altogether */
func initStatus(total int) {
	totalCount = total
}

/* creates new status handle */
func createNewStatus() {
	// increase the global status ID counter
	currentID++

	// refresh the current global instance
	currentStatus = &processingStatus{
		outputWriter: color.Error, // we output to colorized error
		fresh:        true,
		prefix:       fmt.Sprintf("[%d/%d] ", currentID, totalCount),
	}
}

/* closes current status and prints newLine */
func closeCurrentStatus() {
	if currentStatus == nil {
		panic("No status currently open")
	}

	// print line-feed upon closing the status
	fmt.Fprintln(currentStatus.outputWriter, "")
	currentStatus = nil
}

/* printError point of entry: if no status yet exists, just prints as usual */
func printError(err error) {
	errString := redBold(err.Error())

	if currentStatus != nil {
		currentStatus._print(errString, false)
	} else {
		fmt.Fprintf(color.Error, errString)
	}
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

/* used to encode the binary output of the tool */
type encoder interface {
	encode([]byte) string
}

/* generic encoder for decryption mode
outputs ASCII printable range as-is, all other bytes escaped with \x notation */
type escapeString struct{}

func (e escapeString) encode(input []byte) string {
	output := strings.Builder{}
	for _, b := range input {
		str := fmt.Sprintf("%+q", string(b))
		str = str[1 : len(str)-1]
		str = strings.Replace(str, `\"`, `"`, 1)
		output.WriteString(str)
	}
	return output.String()
}

/* general status handle */
type processingStatus struct {
	outputWriter io.Writer // the output to write into
	prefix       string    // prefix of current status
	lineIndex    int       // number of last line that was printed on
	fresh        bool      // indicator if status has already even printed something
	bar          *hackyBar // the dynamically changing, hollywood-style bar inside current status
}

/* starts new hacky bar */
func (p *processingStatus) openBar(outputLen int) {
	/* chose output encoder */
	var encoder encoder
	if *config.encrypt {
		encoder = config.encoder
	} else {
		encoder = escapeString{}
	}

	/* create bar */
	p.bar = &hackyBar{
		status:         p,
		output:         make([]byte, 0, outputLen),
		totalOutputLen: outputLen,
		wg:             sync.WaitGroup{},
		chanOutput:     make(chan int),
		chanReq:        make(chan byte, *config.parallel),
		chanStop:       make(chan byte),
		autoUpdateFreq: time.Second / time.Duration(updateFreq),
		encoder:        encoder,
	}

	// listen for events and reflect the status
	go p.bar.listenAndPrint()
}

/* gracefully shutting down the hackyBar goroutine */
func (p *processingStatus) closeBar() {
	p.bar.chanStop <- 0
	p.bar.wg.Wait()

	// print warning if overflow occurred and stdout was not redirected
	if p.bar.overflow && isTerminal(os.Stdout) {
		printError(errors.New("Output was too wide to fit you terminal. Redirect stdout somewhere to get full output"))
	}
	p.bar = nil
}

/* printing function in context of status */
func (p *processingStatus) _print(s string, sameLine bool) {
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
	fmt.Fprint(p.outputWriter, builder.String())
}

// actions are printed in yellow, on the same line, they are temporary strings
func (p *processingStatus) printAction(s string) {
	p._print(yellow(s), true)
}

// function to be used by external cracker.go to report about yet-another-plaintext-byte revealed
func (p *processingStatus) fetchOutputByte(b byte) {
	p.bar.chanOutput <- int(b)
}

// replaces last fetched bytes with new ones form input byte slice */
func (p *processingStatus) replaceLastFetchedBytes(rep []byte) {
	/* send control magic number */
	p.bar.chanOutput <- ccREPLACEBYTES + len(rep)

	/* send replacement bytes */
	for _, b := range rep {
		p.fetchOutputByte(b)
	}

}
