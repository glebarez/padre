package probe

import (
	"unicode"

	"github.com/glebarez/padre/pkg/client"
)

// ResponseFingerprint ...
type ResponseFingerprint struct {
	StatusCode int
	Lines      int
	Words      int
}

// GetResponseFingerprint - scrape fingerprint form http response
func GetResponseFingerprint(resp *client.Response) (*ResponseFingerprint, error) {
	return &ResponseFingerprint{
		StatusCode: resp.StatusCode,
		Lines:      countLines(resp.Body),
		Words:      countWords(resp.Body),
	}, nil
}

// helper: count number of lines in input
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

// helper: count number of lines in input
func countWords(input []byte) int {
	inWord, count := false, 0
	for _, r := range string(input) {
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}
