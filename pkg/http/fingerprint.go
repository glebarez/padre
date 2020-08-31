package http

import (
	"net/http"
	"unicode"
)

// ResponseFingerprint ...
type ResponseFingerprint struct {
	StatusCode int
	Lines      int
	Words      int
}

// GetResponseFingerprint - scrape fingerprint form http response
func GetResponseFingerprint(resp *http.Response, body []byte) (*ResponseFingerprint, error) {
	return &fingerprint{
		StatusCode: resp.StatusCode,
		Lines:      countLines(body),
		Words:      countWords(body),
	}, nil
}

// count number of lines in input
func countLines(input []byte) int {
	if len(input) == 0 {
		return 0
	}
	count := 1
	for _, b := range input {
		if b == '\n' {
			count++
		}
	}
	return count
}

// count number of words in input
func countWords(input []byte) int {
	inWord, count := false, 0
	for _, r := range string(input) {
		if unicode.IsSpace(r) {
			inWord = false
		} else if inWord == false {
			inWord = true
			count++
		}
	}
	return count
}
