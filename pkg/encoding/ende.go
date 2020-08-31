package encoding

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/glebarez/padre/pkg/util"
)

// DecodeError ...
type DecodeError string

func (e DecodeError) Error() string { return string(e) }

// EncoderDecoder - performs encoding/decoding
type EncoderDecoder interface {
	EncodeToString([]byte) string
	DecodeString(string) ([]byte, error)
}

/* wrapper for encoderDecoder with characters replacements */
type encDecWithReplacer struct {
	ed                     EncoderDecoder
	replacerAfterEncoding  *strings.Replacer
	replacerBeforeDecoding *strings.Replacer
}

// encode with replacement
func (r encDecWithReplacer) EncodeToString(input []byte) string {
	encoded := r.ed.EncodeToString(input)
	return r.replacerAfterEncoding.Replace(encoded)
}

// decode with replacement
func (r encDecWithReplacer) DecodeString(input string) ([]byte, error) {
	encoded := r.replacerBeforeDecoding.Replace(input)
	decoded, err := r.ed.DecodeString(encoded)
	if err != nil {
		return nil, DecodeError(fmt.Sprintf("Decode error: %s", err))
		// errors.NewErrWithHints(fmt.Errorf("decode error: %w", err), hint.checkEncoding, hint.checkInput)
	}
	return decoded, nil
}

// wrapper creator
func wrapEncoderDecoder(ed EncoderDecoder, replacements string) EncoderDecoder {
	return &encDecWithReplacer{
		ed:                     ed,
		replacerAfterEncoding:  strings.NewReplacer(strings.Split(replacements, "")...),
		replacerBeforeDecoding: strings.NewReplacer(strings.Split(util.ReverseString(replacements), "")...),
	}
}

// lowercase hex encoder/decoder
type lhexWrap struct{}

func (h lhexWrap) EncodeToString(input []byte) string {
	return hex.EncodeToString(input)
}

func (h lhexWrap) DecodeString(input string) ([]byte, error) {
	return hex.DecodeString(input)
}

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

// constructor for base64
func createB64encDec(replacements string) EncoderDecoder {
	return wrapEncoderDecoder(base64.StdEncoding, replacements)
}

// constructor for lowercase hex
func createLHEXencDec(replacements string) EncoderDecoder {
	return wrapEncoderDecoder(lhexWrap{}, replacements)
}
