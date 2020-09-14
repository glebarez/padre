package client

import (
	"context"
	"sync"
)

// equals to 2**8, since we're testing every possible value of a byte
const probeCount = 256

// ProbeResult - result of probe
type ProbeResult struct {
	Byte     byte
	Response *Response
	Err      error
}

// SendProbes -  given a chunk of bytes, place every possible byte-value at specified position.
// These probes are sent concurrently over HTTP.
// The results will be written into chanResult channel
func (client *Client) SendProbes(ctx context.Context, chunk []byte, pos int, chanResult chan *ProbeResult) {
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
						chanResult <- &ProbeResult{
							Byte: b,
							Err:  err,
						}
					} else {
						// send response
						chanResult <- &ProbeResult{
							Byte:     b,
							Response: resp,
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
