package encoder

import "encoding/hex"

// lowercase hex encoder/decoder
type lhexEncoder struct{}

func (h *lhexEncoder) EncodeToString(input []byte) string {
	return hex.EncodeToString(input)
}

func (h *lhexEncoder) DecodeString(input string) ([]byte, error) {
	return hex.DecodeString(input)
}
