package main

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const (
	updateFreq = 13 // status update frequency (times/sec)
)

var (
	currentStatus *processingStatus // currently active status (we use singleton for now)
	totalCount    int               // total count of all statuses to be shown
	currentID     int               // id of currently active status
)

func initStatus(total int) {
	totalCount = total
}

// the status handle
type processingStatus struct {
	output    io.Writer
	prefix    string
	lineIndex int
	width     int
	fresh     bool
	bar       *hackyBar
}

// dynamically changing hacky bar
type hackyBar struct {
	// parent
	status *processingStatus

	// plaintext length info
	plainLen        int    // length of overall plain text do deciphered, this is needed for proper formatting
	decipheredPlain string // plain, deciphered so far
	decipheredCount int    // the length of deciphered so far plain, needed because string above may contain escape sequences, thus does not reflect real deciphered length

	// async communications
	chanPlain chan byte      // chan through which real-time communcation about deciphered bytes occures
	chanStop  chan byte      // used to tell async thread that it's time to die...
	wg        sync.WaitGroup // this wg is used to wait till async thread dies, upon closing the bar

	// RPS calculation
	start        time.Time // the time of first request made, needed to properly calculate RPS
	requestsMade int       // total requests made, needed to calculate RPS
	rps          int       // RPS
	chanReq      chan byte

	// the output properties
	autoUpdateFreq time.Duration // interval at which the bar must be updated
}

/* creates new status handle */
func createNewStatus() {
	// increase the ID
	currentID++

	// refresh the current instance
	currentStatus = &processingStatus{
		output: color.Error, // we output to colorized error
		fresh:  true,
		prefix: fmt.Sprintf("[%d/%d]", currentID, totalCount),
	}
}

func (p *hackyBar) listenAndPrint() {
	lastPrint := time.Now()
	stop := false
	p.wg.Add(1)

	for {
		select {
		// collect another revealed byte of plaintext
		case b := <-p.chanPlain:
			p.decipheredCount++
			p.decipheredPlain = escapeChar(b) + p.decipheredPlain

		// another HTTP request was made, count it
		case <-p.chanReq:
			if p.requestsMade == 0 {
				p.start = time.Now()
			}

			p.requestsMade++

			secsPassed := int(time.Since(p.start).Seconds())
			if secsPassed > 0 {
				p.rps = p.requestsMade / int(secsPassed)
			}

		// it's time to stop
		case <-p.chanStop:
			stop = true
		}

		// update status line if it's time to
		// and always print most actual state before we exit
		if time.Since(lastPrint) > p.autoUpdateFreq || stop {
			p.status._print(p.buildStatusString(!stop), true)
			lastPrint = time.Now()
		}

		// return if it's time
		if stop {
			p.wg.Done()
			return
		}
	}
}

// build status string
func (p *hackyBar) buildStatusString(hacky bool) string {
	randLen := p.plainLen - p.decipheredCount

	plain := fmt.Sprintf("%s%s", randString(randLen, hacky), greenBold(p.decipheredPlain))

	status := fmt.Sprintf(
		"%80s (%d/%d) | Requests made: %d (%d/sec)",
		plain,
		p.decipheredCount,
		p.plainLen,
		p.requestsMade,
		p.rps)
	return status
}

/* fires a hacky bar */
func (p *processingStatus) openBar(plainLen int) {
	// create bar
	p.bar = &hackyBar{
		status:         p,
		plainLen:       plainLen,
		wg:             sync.WaitGroup{},
		chanPlain:      make(chan byte),
		chanReq:        make(chan byte, *config.parallel),
		chanStop:       make(chan byte),
		autoUpdateFreq: time.Second / time.Duration(updateFreq),
	}

	// listen for events and reflect the status
	go p.bar.listenAndPrint()
}

func (p *processingStatus) closeBar() {
	p.bar.chanStop <- 0
	p.bar.wg.Wait()
	p.bar = nil

	// print new line
	p._print("", false)
}

func (p *processingStatus) _print(s string, sameLine bool) {
	// after first print, currentStatus will become unfresh
	defer func() {
		if p.fresh {
			p.fresh = false
		}
	}()

	// create builder for efficiency
	builder := &strings.Builder{}
	builder.Grow(p.width)

	// if same line, prepent with caret return
	if sameLine {
		builder.WriteByte('\r')
	} else {
		// well, applying newLine logic to fresh instance is not necessary, really
		// otherwise we end up having a blank line
		if !p.fresh {
			p.lineIndex++
			builder.WriteByte('\n')
		}
	}

	// add prefix only if it's the first line of current status
	if p.lineIndex == 0 {
		builder.WriteString(cyanBold(p.prefix))
		builder.WriteByte(' ')
	}

	// add the input payload
	builder.WriteString(s)

	// output finally
	fmt.Fprint(p.output, builder.String())
}

// actions are printed in the same line, they are temporary strings
func (p *processingStatus) printAction(s string) {
	p._print(yellow(s), true)
}

// a single printError point of access: if no status is yet exists, just print
func printError(err error) {
	errString := redBold(err.Error()) + "\n"

	if currentStatus != nil {
		currentStatus._print(errString, false)
	} else {
		fmt.Fprintf(color.Error, errString)
	}
}

// function to use by external cracker to report about yet-another-plaintext-byte cracked
func (p *processingStatus) reportPlainByte(b byte) {
	p.bar.chanPlain <- b
}

// function to use by external http client that yet-another requiest was made
func (p *processingStatus) reportHTTPRequest() {
	// http client can make requets outside of bar scope (e.g. pre-flight checks)
	if p.bar != nil {
		p.bar.chanReq <- 1
	}

}
