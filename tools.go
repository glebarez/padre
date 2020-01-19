package main

import (
	"encoding/base64"
	"log"
)

func decode(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func decipherChunk(chunk []byte) ([]byte, error) {
	/* okay, we're now very close to the Padding Oracle core technique */
	plainText := make([]byte, blockLen)

	// we start with the last byte of first block
	// and repeat the same procedure for every byte in that block
	for pos := blockLen - 1; pos >= 0; pos-- {
		originalByte := chunk[pos]

		// try every single byte value, except the original one
		found := false
		foundByte := byte(0)

		for i := 0; i < 0xff; i++ {
			// we skip the original byte, since we know it's okay
			if byte(i) == originalByte {
				continue
			}

			// set the byte to tampered one
			chunk[pos] = byte(i)

			// test for padding oracle
			paddingError, err := isPaddingError(chunk)
			if err != nil {
				log.Fatal(err)
			}

			// if it is an Error, skip to the next
			if paddingError {
				continue
			} else {
				// but if we found the valid tampered cipher, it's good news!
				found = true
				foundByte = byte(i)
				break
			}
		}

		// okay, let's check how we ended up inside the loop above
		if !found {
			// well, seems like the only valid padding is the original byte
			// that can happen if we just hit the original padding in our tampering attempts
			// so, let it be, but also check that to be sure
			chunk[pos] = originalByte
			paddingError, err := isPaddingError(chunk)
			if err != nil {
				log.Fatal(err)
			}
			if paddingError {
				log.Fatal("Failed to decrypt, not a byte got in without padding error. The root cause migth be in Unpadding implementation on the server side!")
			}

			log.Printf("Only original byte fits at position %d", pos)
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
		then OPT = OCB ^ TCB ^ TPT
		*/
		currPaddingValue := byte(16 - pos)
		plainText[pos] = foundByte ^ originalByte ^ currPaddingValue

		/* good! we're done with this byte, do we forget something?
		yes, we actually need to repair the padding for the next shot
		e.g. we need to adjust the already tapered bytes of first */
		nextPaddingValue := currPaddingValue + 1
		adjustingValue := currPaddingValue ^ nextPaddingValue
		for i := pos; i < blockLen; i++ {
			chunk[i] ^= adjustingValue
		}
	}

	// now we're ready to return our revealed plaintext
	return plainText, nil
}
