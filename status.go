package main

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/fatih/color"
)

var currentStatus *processingStatus
var plainWidth = 80

/* this has to be justified */
type processingStatus struct {
	plainLen        int
	decipheredPlain string
	decipheredCount int
	chanPlain       chan byte
	wg              sync.WaitGroup
	start           time.Time
	requestsMade    int
	rps             int
	chanReq         chan byte
	output          io.Writer
	autoUpdateFreq  time.Duration
	chanStop        chan byte
	prefix          string
}

func createStatus(current, total int) *processingStatus {
	status := &processingStatus{
		output:         color.Error,
		autoUpdateFreq: time.Second / 10,
		prefix:         fmt.Sprintf("[%d/%d]", current, total),
	}

	currentStatus = status
	return status
}

// count request and adjust statistics
func (p *processingStatus) countRequest() {
	if p.requestsMade == 0 {
		p.start = time.Now()
	}
	p.requestsMade++
	secsPassed := int(time.Since(p.start).Seconds())
	if secsPassed > 0 {
		p.rps = p.requestsMade / int(secsPassed)
	}
}

// build status string
func (p *processingStatus) buildStatusString() string {
	randLen := p.plainLen - p.decipheredCount

	plain := fmt.Sprintf("%s%s", randString(randLen), greenBold(p.decipheredPlain))

	status := fmt.Sprintf(
		"%80s (%d/%d) | Requests made: %d (%d/sec)",
		plain,
		// randString(randLen),
		// greenBold(p.decipheredPlain),
		p.decipheredCount,
		p.plainLen,
		p.requestsMade,
		p.rps)
	return status
}

func (p *processingStatus) prefixed(s string) string {
	return fmt.Sprintf("%s %s", cyanBold(p.prefix), s)
}

func (p *processingStatus) printSameLine(s string) {
	fmt.Fprintf(p.output, "\r%s", p.prefixed(s))
}

func (p *processingStatus) printNewLine(s string) {
	fmt.Fprintf(p.output, "\n%s", s)
}

func (p *processingStatus) printAppend(s string) {
	fmt.Fprintf(p.output, "%s", s)
}

func (p *processingStatus) printAction(s string) {
	p.printSameLine(yellowBold(s))
}

func (p *processingStatus) finishStatusBar() {
	// wait untill all the plaintext is recieved
	p.wg.Wait()

	// stop thread  to avoid goroutine leak
	p.chanStop <- 0

	// print the final status string
	p.printSameLine(p.buildStatusString() + "\n")
}

func (p *processingStatus) error(err error) {
	// stop thread if it is running, to avoid goroutine leak
	if p.chanStop != nil {
		p.chanStop <- 0

		// print the current status without hacky stuff
		hacky = false
		p.printSameLine(p.buildStatusString())
		hacky = true
	}

	// print the error which caused the abort
	p.printNewLine(red(err.Error()) + "\n")
}

func (p *processingStatus) startStatusBar(plainLen int) {
	// prepare all the stuff for async work
	p.plainLen = plainLen
	p.wg = sync.WaitGroup{}
	p.wg.Add(plainLen)

	p.chanPlain = make(chan byte)
	p.chanReq = make(chan byte, *config.parallel)
	p.chanStop = make(chan byte)

	// start loop in separate thread
	go func() {
		var lastPrint = time.Now()
		for {
			select {
			case b := <-p.chanPlain:
				// correct the plain text info
				p.decipheredCount++
				p.decipheredPlain = escapeChar(b) + p.decipheredPlain
				p.wg.Done()

			case <-p.chanReq:
				p.countRequest()

			case <-p.chanStop:
				return
			}

			// update status line if it's time
			if time.Since(lastPrint) > p.autoUpdateFreq {
				p.printSameLine(p.buildStatusString())
				lastPrint = time.Now()
			}
		}
	}()
}
