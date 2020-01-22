package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
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
	output          io.WriteCloser
	autoUpdateFreq  int
}

func createStatus(plainLen int) *processingStatus {
	status := &processingStatus{
		plainLen:  plainLen,
		wg:        sync.WaitGroup{},
		output:    os.Stderr,
		chanPlain: make(chan byte),
		chanReq:   make(chan byte, parallel),
	}
	status.wg.Add(plainLen)

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
	plain := fmt.Sprintf("%s%s", randStringBytes(randLen), string(p.decipheredPlain))
	status := fmt.Sprintf(
		"%s (%d/%d) | Requests made: %d (%d/sec)",
		plain,
		p.decipheredCount,
		p.plainLen,
		p.requestsMade,
		p.rps)
	return status
}

func (p *processingStatus) waitFinish() {
	p.wg.Wait()
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

func (p *processingStatus) close() {
	// wait untill all the plaintext is recieved
	p.wg.Wait()

	// print the final status string
	p.printSameLine(p.buildStatusString())

	// print newline
	p.printNewLine("")
}

func (p *processingStatus) startStatusBar() {
	// get ticker
	ticker := time.NewTicker(time.Second / time.Duration(p.autoUpdateFreq))

	// start loop in separate thread
	go func() {
		for {
			select {
			case <-ticker.C:
				p.printSameLine(p.buildStatusString())
			case b, ok := <-p.chanPlain:
				// correct the plain text info
				p.decipheredCount++
				p.decipheredPlain = escapeChar(b) + p.decipheredPlain
				p.wg.Done()

				// exit this goroutine when channel is closed
				if !ok {
					return
				}
			case <-p.chanReq:
				p.countRequest()
			}
		}
	}()
}
