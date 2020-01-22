package main

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/fatih/color"
)

var currentStatus *processingStatus

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
	currentCipher   int
	totalCiphers    int
	output          io.Writer
	autoUpdateFreq  int
	chanStop        chan byte
	prefix          string
}

func createStatus() *processingStatus {
	status := &processingStatus{
		output:         color.Error,
		autoUpdateFreq: 10,
	}

	currentStatus = status
	return status
}

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

func (p *processingStatus) buildStatusString() string {
	randLen := p.plainLen - p.decipheredCount

	status := fmt.Sprintf(
		"%s%s (%d/%d) | Requests made: %d (%d/sec)",
		randString(randLen),
		greenBold(p.decipheredPlain),
		p.decipheredCount,
		p.plainLen,
		p.requestsMade,
		p.rps)
	return status
}

func (p *processingStatus) printSameLine(s string) {
	fmt.Fprintf(p.output, "\r%s", s)
}

func (p *processingStatus) printNewLine(s string) {
	fmt.Fprintf(p.output, "\n%s", s)
}

func (p *processingStatus) printAppend(s string) {
	fmt.Fprintf(p.output, "%s", s)
}

func (p *processingStatus) finishStatusBar() {
	// wait untill all the plaintext is recieved
	p.wg.Wait()

	// stop thread  to avoid goroutine leak
	p.chanStop <- 0

	// print the final status string
	p.printSameLine(p.buildStatusString())

	// print newline
	p.printNewLine("")
}

func (p *processingStatus) error(err error) {
	// stop thread if it is running, to avoid goroutine leak
	if p.chanStop != nil {
		p.chanStop <- 0
	}

	// print the current status without hacky stuff
	hacky = false
	p.printSameLine(p.buildStatusString())

	// print the error which caused the abort
	p.printNewLine(red(err.Error()))
	p.printNewLine("")
}

func (p *processingStatus) startStatusBar(plainLen int) {
	// prepare all the stuff for async work
	p.plainLen = plainLen
	p.wg = sync.WaitGroup{}
	p.wg.Add(plainLen)

	p.chanPlain = make(chan byte)
	p.chanReq = make(chan byte, parallel)
	p.chanStop = make(chan byte)

	// get ticker
	ticker := time.NewTicker(time.Second / time.Duration(p.autoUpdateFreq))

	// start loop in separate thread
	go func() {
		for {
			select {
			case <-ticker.C:
				p.printSameLine(p.buildStatusString())

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
		}
	}()
}
