package probe

import (
	"context"
	"sync"

	"github.com/glebarez/padre/pkg/client"
)

// equals to 2**8, since we're testing every possible value of a byte
const probeCount = 256

type probeResult struct {
	probe    byte
	response *client.Response
	err      error
}

/* given a chunk of bytes, place every possible byte-value at specified position.
these probes are sent concurrently over HTTP
the results of probeFunc will be written into chanResult channel */
func sendProbes(ctx context.Context, client *client.Client, chunk []byte, pos int, chanResult chan *probeResult) {
	// send byte values into this
	chanIn := make(chan byte, probeCount)

	/* run workers */
	wg := sync.WaitGroup{}
	for i := 0; i < client.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// copy chunk to produce local concurrent-safe copy
			chunkCopy := copySlice(chunk)

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
					resp, err := client.DoRequest(ctx, chunkCopy)
					if ctx.Err() == context.Canceled {
						return
					}

					if err != nil {
						// error during HTTP request
						chanResult <- &probeResult{
							probe: b,
							err:   err,
						}
					} else {
						// send response
						chanResult <- &probeResult{
							probe:    b,
							response: resp,
						}
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
}
