package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

/* Encoders/Decoders */

// interface
type encoderDecoder interface {
	EncodeToString([]byte) string
	DecodeString(string) ([]byte, error)
}

/* wrapper for encoderDecoder with characters replacements */
type encDecWithReplacer struct {
	ed                     encoderDecoder
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
		return nil, newErrWithHints(fmt.Errorf("decode error: %w", err), hint.checkEncoding, hint.checkInput)
	}
	return decoded, nil
}

// wrapper creator
func wrapEncoderDecoder(ed encoderDecoder, replacements string) encoderDecoder {
	return &encDecWithReplacer{
		ed:                     ed,
		replacerAfterEncoding:  strings.NewReplacer(strings.Split(replacements, "")...),
		replacerBeforeDecoding: strings.NewReplacer(strings.Split(reverseString(replacements), "")...),
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
				die(err)
			}
		} else {
			_, err := output.WriteString(fmt.Sprintf("\\x%02x", b))
			if err != nil {
				die(err)
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
func createB64encDec(replacements string) encoderDecoder {
	return wrapEncoderDecoder(base64.StdEncoding, replacements)
}

// constructor for lowercase hex
func createLHEXencDec(replacements string) encoderDecoder {
	return wrapEncoderDecoder(lhexWrap{}, replacements)
}
