package main

import (
	"fmt"
	"io"
	"math/rand"
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

// global init, to tell how many ciphers if there to crack
func initStatus(total int) {
	totalCount = total
}

/* generates string that represents the yet-unrevealed portion of plaintext
if in hacky mode, will produce random characters */
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

/* escapes unprintable characters without quoting them on sides */
func escapeChar(char byte) string {
	str := fmt.Sprintf("%+q", string(char))
	str = str[1 : len(str)-1]
	return strings.Replace(str, `\"`, `"`, 1)
}

/* general status handle */
type processingStatus struct {
	output    io.Writer // the output to write into
	prefix    string    // prefix of current status
	lineIndex int       // number of last line that was printed on
	fresh     bool      // indicator if status has already even printed something
	bar       *hackyBar // the dynamically changing, hollywood-stype bar
}

/* dynamically changing hacky bar */
type hackyBar struct {
	// parent
	status *processingStatus

	// plaintext length info
	plainLen        int    // length of overall plain text to be deciphered, this is needed for proper formatting
	decipheredPlain string // plaintext, deciphered so far
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
	// increase the global status ID counter
	currentID++

	// refresh the current global instance
	currentStatus = &processingStatus{
		output: color.Error, // we output to colorized error
		fresh:  true,
		prefix: fmt.Sprintf("[%d/%d] ", currentID, totalCount),
	}
}

func closeStatus() {
	if currentStatus == nil {
		panic("No status currently open")
	}

	fmt.Fprintln(currentStatus.output, "")
	currentStatus = nil
}

/* background goroutine, which collects information about process and progress
and then prints out the info in hackyBar */
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

		//  exit if it's time
		if stop {
			p.wg.Done() // just to let know those waiting for you to die
			return
		}
	}
}

// build status string
func (p *hackyBar) buildStatusString(hacky bool) string {
	// statistics part of output
	stats := fmt.Sprintf(
		"(%d/%d) | reqs: %d (%d/sec)",
		p.decipheredCount,
		p.plainLen,
		p.requestsMade,
		p.rps)

	// plain part
	unkLen := p.plainLen - p.decipheredCount
	plain := unknownString(unkLen, hacky) + p.decipheredPlain

	// get plain part to be showed (taking into account available terminal width)
	leftRoom := config.termWidth - (len(p.status.prefix) + len(plain) + len(stats) + 1) // + 1 for space between stats and plain

	if leftRoom < 0 {
		/* we actually get to output more than terminal is gonna take, so let's cut things out */
		cutUpTo := len(plain) + leftRoom - 3 // leftRoom is negative
		if cutUpTo < 0 {
			panic("Your terminal is to narrow! Use a real one")
		}
		plain = plain[:cutUpTo] + `...`
	} else {
		plain += strings.Repeat(" ", leftRoom)
	}

	// make the deciphered part colorized
	if len(plain) > unkLen {
		plain = plain[:unkLen] + hiGreenBold(plain[unkLen:])
	}

	// build the final string
	return fmt.Sprintf("%s %s", plain, stats)
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

/* gracefully shutting down the hackyBar goroutine */
func (p *processingStatus) closeBar() {
	p.bar.chanStop <- 0
	p.bar.wg.Wait()
	p.bar = nil
}

/* printing function in context of status */
func (p *processingStatus) _print(s string, sameLine bool) {
	// after first print, currentStatus will become unfresh
	defer func() {
		if p.fresh {
			p.fresh = false
		}
	}()

	// create builder for efficiency
	builder := &strings.Builder{}
	builder.Grow(config.termWidth)

	// if same line, prepend with caret return
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
	} else {
		// othewise, just put spaces for nice indent
		builder.WriteString(strings.Repeat(" ", len(p.prefix)))
	}

	builder.WriteString(s)

	// output finally
	fmt.Fprint(p.output, builder.String())
}

// actions are printed in the same line, they are temporary strings
func (p *processingStatus) printAction(s string) {
	p._print(yellow(s), true)
}

// a single printError point of entry: if no status yet exists, just prints as usual
func printError(err error) {
	errString := redBold(err.Error())

	if currentStatus != nil {
		currentStatus._print(errString, false)
	} else {
		fmt.Fprintf(color.Error, errString)
	}
}

// function to be used by external cracker to report about yet-another-plaintext-byte revealed
func (p *processingStatus) reportPlainByte(b byte) {
	p.bar.chanPlain <- b
}

// function to be used by external http client to report that yet-another requiest was made
func reportHTTPRequest() {
	if currentStatus == nil {
		return
	}

	// http client can make requets outside of bar scope (e.g. pre-flight checks), those do not count
	if currentStatus.bar != nil {
		currentStatus.bar.chanReq <- 1
	}
}
