package main

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
)

/* Encoders/Decoders */

type encoderDecoder interface {
	encode([]byte) string
	decode(string) ([]byte, error)
}

type b64 struct {
	replacerAfterEncoding  *strings.Replacer
	replacerBeforeDecoding *strings.Replacer
}

type lhex struct {
	replacerAfterEncoding  *strings.Replacer
	replacerBeforeDecoding *strings.Replacer
}

func reverseString(in string) string {
	out := strings.Builder{}
	for i := len(in) - 1; i >= 0; i-- {
		out.WriteByte(in[i])
	}
	return out.String()
}

func createBase64EncoderDecoder(replaceAfterEncoding string) encoderDecoder {
	ende := &b64{}

	// create replacers
	ende.replacerAfterEncoding = strings.NewReplacer(strings.Split(replaceAfterEncoding, "")...)
	ende.replacerBeforeDecoding = strings.NewReplacer(strings.Split(reverseString(replaceAfterEncoding), "")...)
	return ende
}

func createLowerHexEncoderDecoder(replaceAfterEncoding string) encoderDecoder {
	ende := &lhex{}

	// create replacers
	ende.replacerAfterEncoding = strings.NewReplacer(strings.Split(replaceAfterEncoding, "")...)
	ende.replacerBeforeDecoding = strings.NewReplacer(strings.Split(reverseString(replaceAfterEncoding), "")...)
	return ende
}

func (b b64) decode(in string) ([]byte, error) {
	// apply replacer
	in = b.replacerBeforeDecoding.Replace(in)

	// decode base64
	out, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (b b64) encode(in []byte) string {
	out := base64.StdEncoding.EncodeToString(in)

	// apply replacer
	return b.replacerAfterEncoding.Replace(out)
}

func (l lhex) decode(in string) ([]byte, error) {
	out, err := hex.DecodeString(in)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (l lhex) encode(in []byte) string {
	out := hex.EncodeToString(in)
	return l.replacerAfterEncoding.Replace(strings.ToLower(out))
}
