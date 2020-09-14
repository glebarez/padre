package encoder

import (
	"fmt"
	"strings"
)

// ASCII encoder
type asciiEncoder struct{}

// escapes non standard ASCII with \x notation
func (e asciiEncoder) EncodeToString(input []byte) string {
	output := strings.Builder{}
	for _, b := range input {
		if b >= 32 && b <= 127 {
			// ascii printable
			err := output.WriteByte(b)
			if err != nil {
				panic(err)
			}
		} else {
			_, err := output.WriteString(fmt.Sprintf("\\x%02x", b))
			if err != nil {
				panic(err)
			}
		}
	}
	return output.String()
}

// ... just to comply with interface
func (e asciiEncoder) DecodeString(input string) ([]byte, error) {
	panic("Not implemented")
}
