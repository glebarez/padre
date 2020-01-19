package main

import (
	"fmt"
	"log"
)

const blockLen = 16

var baseURL = "http://34.74.105.127/2edee56f24/?post=%s"
var cipherEncoded = "9P3GWr4PEd8F1e2AR-HopshWrUfNiLWFX2gkVlI1BDLlnAQabyqyExdOXjoSe75Cf0PYFufmqxWHxfq85JwzQ8LOGg2rrVZyuTJ4GElK8ENjo4S3bBcl7N1BEagNrHh1mEROsXltQp!WruUmV0t9NGMEYj1CHyq895JzcxSveF5YwAWmw4mts5xU4nGVPMqpvU0YR3T5SKextRb24Rd77w~~"

func main() {
	/* hey there, here we go again, fresh and clean
	in this chapter, we are going to implement a Padding Oracle exploit
	starting with really simple stuff, we eventually will produce a solid product
	with neat & nice parallelization, progress bars, and hollywood-stye looking hack! */

	// usually we are given an initial, valid cipher, tampering on which, we discover the plaintext! get ready!

	// we decode it into bytes, so we can tamper it at that byte level
	cipher, err := decode(cipherEncoded)
	if err != nil {
		log.Fatal(err)
	}

	// we also need to check the overall cipher length complies with blockLen
	// as this is crucial to further logic
	if len(cipher)%blockLen != 0 {
		log.Fatal("Cipher len is bad")
	}
	blockCount := len(cipher)/blockLen - 1

	/* now, we gonna tamper at every block separately,
	thus we need to split up the whole payload into blockSize*2 sized chunks
	why that size?
	- first half - the bytes we gonna tamper on
	- second half - the bytes that will produce the padding error */
	cipherChunks := make([][]byte, blockCount)
	for i := 0; i < blockCount; i++ {
		cipherChunks[i] = make([]byte, blockLen*2)
		copy(cipherChunks[i], cipher[i*blockLen:(i+2)*blockLen])
	}

	/* so far good
	now, it's time to write the code which can decipher a single block
	this way, we feed it every cipherChunk, and get the plaintext result back! */
	// create container for a final plaintext
	plainText := make([]byte, len(cipher)-blockLen)

	// decode every cipher chunk and fill-in the relevant plaintext positions
	for i, cipherChunk := range cipherChunks {
		plainChunk, err := decipherChunk(cipherChunk)
		if err != nil {
			log.Fatal(err)
		}
		copy(plainText[i*16:(i+1)*16], plainChunk)
	}

	// that's it!
	fmt.Println(string(plainText))
}
