package main

import (
	"fmt"
	"log"
	//	_ "net/http/pprof"
)

const blockLen = 16

var parallel = 20
var baseURL = "http://localhost:5000/decrypt?cipher=%s" //"http://34.74.105.127/2edee56f24/?post=%s"
var cipherEncoded = "jigNcuWcyzd8QB7E/fm7peYSX9gnh6/gYG5Hmy/Bz7IVHVUM1hFyoCjPREV5efzK"
var paddingError = "IncorrectPadding"

func main() {

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
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
	outChan := hollyHack(len(plainText))
	// for i, cipherChunk := range cipherChunks {
	for i := len(cipherChunks) - 1; i >= 0; i-- {
		plainChunk, err := decipherChunk(cipherChunks[i], outChan)
		if err != nil {
			log.Fatal(err)
		}
		copy(plainText[i*16:(i+1)*16], plainChunk)
	}

	// that's it!
	fmt.Printf("\r%s", string(plainText))
}
