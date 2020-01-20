package main

import (
	"fmt"
	"math/rand"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func hollyHack(plainLen int) chan byte {
	// create channel for external writer
	input := make(chan byte, 100)

	// get ticker with update interval we need to print the shit
	t := time.NewTicker(time.Second / 10)

	// here, the real plaintext will be stored as it comes to the channel
	realDeal := make([]byte, 0, plainLen)

	out := func() {
		fmt.Printf("\r%s%s", randStringBytes(plainLen-len(realDeal)), string(realDeal))
	}

	// start printing shit, sometime replacing with real deal
	go func() {

		for {
			select {
			case <-t.C:
				// print the shit
				out()
			case b, ok := <-input:
				// return when channel is closed
				if !ok {
					out()
					return
				}

				// correct the real deal
				realDeal = append([]byte{b}, realDeal...)
			}
		}
	}()
	return input
}
