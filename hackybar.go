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
			it that case we don't need to noise the unprocessed part of output with random noise */
			p.status._print(p.buildStatusString(!stop), true)
			lastPrint = time.Now()
		}

		/* exit when stop requested */
		if stop {
			p.wg.Done() // just to let know those waiting for you to die
			return
		}
	}
}

/* constucts full status string to be displayed */
func (p *hackyBar) buildStatusString(hacky bool) string {
	/* generate stats info */
	stats := fmt.Sprintf(
		"(%d/%d) | reqs: %d (%d/sec)", len(p.output), p.totalOutputLen, p.requestsMade, p.rps)

	/* calculate the length of hacky string - the part of output that is not yet calculated */
	unkLen := p.totalOutputLen - len(p.output)

	/* in case of encryption, we must take length of encoded output */
	if *config.encrypt {
		unkLen = len(p.encoder.encode(make([]byte, unkLen)))
	}

	/* produce final output string for hacky bar */
	data := unknownString(unkLen, hacky) + p.encoder.encode(p.output)

	/* depending on terminal width, we must cut the output part that does not fit */
	leftRoom := config.termWidth - (len(p.status.prefix) + len(data) + len(stats) + 1) // + 1 for space between stats and plain
	if leftRoom < 0 {
		/* we actually get to output more than terminal is gonna take, so let's cut things out */
		cutUpTo := len(data) + leftRoom - 3 // leftRoom is negative, -3 is for tree dots (see below)
		if cutUpTo < 0 {
			panic("Your terminal is to narrow! Use a real one")
		}
		data = data[:cutUpTo] + `...`
	} else {
		/* when room is enough to output everything
		just add some padding, so that stats will always be placed in the right edge of terminal */
		data += strings.Repeat(" ", leftRoom)
	}

	/* make the already computed part of output data colorized green */
	if len(data) > unkLen {
		data = data[:unkLen] + hiGreenBold(data[unkLen:])
	}

	/* build the final string */
	return fmt.Sprintf("%s %s", data, stats)
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
