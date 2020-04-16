package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

/*
Hacky-Bar is the dynamically changing bar in status line.
The bar reflects current state of output calculation.
Apart from currently calculated part of output, it also shows yet-unknown part as a random mix of ASCII characters.
This bar is designed to be fun and fast-changing.
It also shows HTTP-client performance in real-time, such as: total http requests sent, average RPS
NOTE: the hacky-bar is a part of status (see status.go) and cannot be used separately
*/
type hackyBar struct {
	// parent status instance
	status *processingStatus

	// output info
	totalOutputLen int     // total length of expected output, needed for progress tracking
	output         []byte  // so-far computed output result
	encoder        encoder // encoder for the byte-output
	overflow       bool    // flag: terminal width overflowed, data was too wide
	// async communications
	chanOutput chan int       // delivering every byte of output via this channel, int is used to distingush control magic-numbers from true output bytes
	chanStop   chan byte      // used to send a stop-signal to goroutine
	wg         sync.WaitGroup // used to wait for gracefull exit after stop signal sent

	// RPS calculation
	start        time.Time // the time of first request made, needed to properly calculate RPS
	requestsMade int       // total requests made, needed to calculate RPS
	rps          int       // RPS
	chanReq      chan byte

	// the output properties
	autoUpdateFreq time.Duration // interval at which the bar must be updated
}

/* designed to be run as goroutine.
collects information about current progress and then prints the info in hackyBar */
func (p *hackyBar) listenAndPrint() {
	lastPrint := time.Now() // time since last print
	stop := false           // flag: stop requested
	p.wg.Add(1)

	/* listen for incoming events */
	for {
		select {
		/* yet another output byte produced */
		case b := <-p.chanOutput:
			// normal output byte
			if b >= 0 && b <= 255 {
				p.output = append([]byte{byte(b)}, p.output...) //TODO: optimize this
			} else if b >= ccREPLACEBYTES {
				// special control character: replace lastly delivered bytes with with new ones
				for i := 0; i < b-ccREPLACEBYTES; i++ {
					b := byte(<-p.chanOutput)
					if len(p.output) > i {
						p.output[i] = b
					} else {
						p.output = append(p.output, b)
					}
				}
			}

		/* yet another HTTP request was made. Update stats */
		case <-p.chanReq:
			if p.requestsMade == 0 {
				p.start = time.Now()
			}

			p.requestsMade++

			secsPassed := int(time.Since(p.start).Seconds())
			if secsPassed > 0 {
				p.rps = p.requestsMade / int(secsPassed)
			}

		/* stop requested */
		case <-p.chanStop:
			stop = true
		}

		/* output actual state when:
		- it's time to: counting since last time
		- before stopping */
		if time.Since(lastPrint) > p.autoUpdateFreq || stop {
			/* NOTE, we avoid hacky mode (using !stop),
			this is because stop can be requested when some error happened,
			it that case we don't need to noise the unprocessed part of output with hacky string */
			statusString := p.buildStatusString(!stop)
			p.status._print(statusString, true)
			lastPrint = time.Now()
		}

		/* exit when stop requested */
		if stop {
			p.wg.Done() // just to let know those waiting for you to die
			return
		}
	}
}

/* constructs full status string to be displayed */
func (p *hackyBar) buildStatusString(hacky bool) string {
	/* the hacky-bar string is comprised of following parts |unknownOutput|knownOutput|stats|
	- unknown output is the part of output that is not yet calculated, it is represented as 'hacky' string
	- known output is the part of output that is already calculated, it is represented as output, encoded with *p.encoder
	- stats
	*/

	/* generate unknown output */
	unprocessedLen := p.totalOutputLen - len(p.output)
	if *config.encrypt {
		unprocessedLen = len(p.encoder.encode(make([]byte, unprocessedLen)))
	}
	unknownOutput := unknownString(unprocessedLen, hacky)

	/* generate known output */
	knownOutput := p.encoder.encode(p.output)

	/* generate stats */
	stats := fmt.Sprintf(
		"(%d/%d) | reqs: %d (%d/sec)", len(p.output), p.totalOutputLen, p.requestsMade, p.rps)

	/* get available space in current terminal width */
	availableSpace := config.termWidth - len(p.status.prefix) - len(stats) - 1 // -1 is for the space between output and stats
	if availableSpace < 5 {
		// a general fool-check
		panic("Your terminal is to narrow. Use a real one")
	}

	/* if we have enough space, the logic is simple */
	if availableSpace >= len(unknownOutput)+len(knownOutput) {
		output := unknownOutput + hiGreenBold(knownOutput)

		// pad with spaces to make stats always appear at the right edge of the screen
		output += strings.Repeat(" ", availableSpace-len(unknownOutput)-len(knownOutput))
		return fmt.Sprintf("%s %s", output, stats)
	}

	/* if we made it to here, we need to cut the output to fit into the available space
	the main idea is to choose the split-point - the poisition at which unknown output ends and known output starts */

	// at first, chose at 1/3 of available space
	splitPoint := availableSpace / 3

	// correct if knownOutput is too short yet
	if len(knownOutput) < availableSpace-splitPoint {
		splitPoint = availableSpace - len(knownOutput)
	} else if len(unknownOutput) < splitPoint {
		// correct if unknownOutput is too short
		splitPoint = len(unknownOutput)
	}

	// put ... into the end of knownOutput if it's too long
	if len(knownOutput) > availableSpace-splitPoint {
		knownOutput = knownOutput[:availableSpace-splitPoint-3] + `...`
		p.overflow = true
	}

	outputString := unknownOutput[:splitPoint] + hiGreenBold(knownOutput)

	/* return the final string */
	return fmt.Sprintf("%s %s", outputString, stats)
}

/* generates string that represents the yet-unknown portion of output
when in 'hacky' mode, will produce random characters form ASCII printable range*/
func unknownString(n int, hacky bool) string {
	b := make([]byte, n)
	for i := range b {

		if hacky {
			b[i] = byte(rand.Intn(126-33) + 33) // byte from ASCII printable range
		} else {
			b[i] = '_'
		}
	}
	return string(b)
}
