package http

import (
	"context"
	"net/http"
	"sync"
)

/* function type that will process probe result, used in sendProbes() */
type probeFunc func(*http.Response, []byte) (interface{}, error)

type probeResult struct {
	probe  byte
	result interface{} // returned from probeFunc
	err    error       // returned from probeFunc
}

/* given a chunk of bytes, place every possible byte-value at specified position.
these probes are sent concurrently over HTTP, the responses will be processed by specified probeFunc
the results of probeFunc will be written into output channel
the channel is returned to the caller, to read from it*/
func sendProbes(ctx context.Context, chunk []byte, pos int, probeFunc probeFunc) chan *probeResult {
	/* how many unique probes will be run.
	equals to 2**8, since we're testing every possible value of a byte */
	const probeCount = 256

	/* communication channels, buffered for:
	- performance
	- to avoid goroutine leak due to deadlocks in edge cases*/
	chanIn := make(chan byte, probeCount)
	chanResult := make(chan *probeResult, probeCount)

	/* run workers */
	wg := sync.WaitGroup{}
	for i := 0; i < *config.parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// copy chunk to produce local concurrent-safe copy
			chunkCopy := sliceCopy(chunk)

			// do the work
			for {
				select {
				case <-ctx.Done():
					// early exit if context is cancelled
					return
				case b, ok := <-chanIn:
					// exit when input channel exhausted
					if !ok {
						return
					}

					// modify byte at given position
					chunkCopy[pos] = b

					// make HTTP request
					resp, body, err := doRequest(ctx, chunkCopy)
					if ctx.Err() == context.Canceled {
						return
					}

					// error during HTTP request
					if err != nil {
						chanResult <- &probeResult{
							probe: b,
							err:   err,
						}
						continue
					}

					// process probe result
					result, err := probeFunc(resp, body)
					chanResult <- &probeResult{
						probe:  b,
						result: result,
						err:    err,
					}
				}
			}
		}()
	}

	/* close output channel when workers are done */
	go func() {
		wg.Wait()
		close(chanResult)
	}()

	/* input generator: every possible byte value */
	go func() {
		for i := 0; i <= 0xff; i++ {
			chanIn <- byte(i)
		}
		close(chanIn)
	}()

	/* deliver output channel to caller */
	return chanResult
}
