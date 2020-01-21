package main

import (
	"fmt"
	"sync"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var chanReq = make(chan byte)

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		//b[i] = letterBytes[rand.Intn(len(letterBytes))]
		b[i] = '_'
	}
	return string(b)
}

type processingStatus struct {
	requestsMade    int
	plainLen        int
	decipheredPlain []byte
	wg              sync.WaitGroup
	start           time.Time
	rps             int
}

func createStatus(plainLen int) *processingStatus {
	status := &processingStatus{
		plainLen:        plainLen,
		decipheredPlain: make([]byte, 0),
		wg:              sync.WaitGroup{},
	}
	status.wg.Add(plainLen)
	return status
}

func (p *processingStatus) countRequest() {
	if p.requestsMade == 0 {
		p.start = time.Now()
	}
	p.requestsMade++
	secsPassed := int(time.Now().Sub(p.start).Seconds())
	if secsPassed > 0 {
		p.rps = p.requestsMade / int(secsPassed)
	}
}

func (p *processingStatus) print(final bool) {
	if final {
		p.wg.Wait()
	}

	randLen := p.plainLen - len(p.decipheredPlain)
	plain := fmt.Sprintf("%s%s", randStringBytes(randLen), string(p.decipheredPlain))
	status := fmt.Sprintf(
		"%+q (%d/%d) | Requests made: %d (%d/sec)",
		plain,
		len(p.decipheredPlain),
		p.plainLen,
		p.requestsMade,
		p.rps)

	fmt.Printf(status)
	if final {
		fmt.Printf("\n")
	} else {
		fmt.Print("\r")
	}
}

func hollyHack(plainLen int) (chan byte, *processingStatus) {
	status := createStatus(plainLen)

	// create channel for external writer
	input := make(chan byte, 100)

	// get ticker with update interval we need to print the shit
	t := time.NewTicker(time.Second / 10)

	// start printing shit, sometime replacing with real deal
	go func() {
		for {
			select {
			case <-t.C:
				status.print(false)
			case b, ok := <-input:
				// correct the real deal
				status.decipheredPlain = append([]byte{b}, status.decipheredPlain...)
				status.wg.Done()

				// return when channel is closed
				if !ok {
					return
				}
			case <-chanReq:
				status.countRequest()
			}
		}
	}()
	return input, status
}
