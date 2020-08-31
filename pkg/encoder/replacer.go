package encoder

import (
	"fmt"
	"strings"

	"github.com/glebarez/padre/pkg/util"
)

/* wrapper for encoderDecoder with characters replacements */
type encoderWithReplacer struct {
	encoder                Encoder
	replacerAfterEncoding  *strings.Replacer
	replacerBeforeDecoding *strings.Replacer
}

// encode with replacement
func (r *encoderWithReplacer) EncodeToString(input []byte) string {
	encoded := r.ed.EncodeToString(input)
	return r.replacerAfterEncoding.Replace(encoded)
}

// decode with replacement
func (r *encoderWithReplacer) DecodeString(input string) ([]byte, error) {
	encoded := r.replacerBeforeDecoding.Replace(input)
	decoded, err := r.encoder.DecodeString(encoded)
	if err != nil {
		return nil, DecodeError(fmt.Sprintf("Decode error: %s", err))
		// errors.NewErrWithHints(fmt.Errorf("decode error: %w", err), hint.checkEncoding, hint.checkInput)
	}
	return decoded, nil
}

// wrapper creator
func newEncDecWithReplacer(encoder Encoder, replacements string) Encoder {
	return &encDecWithReplacer{
		encoder:                encoder,
		replacerAfterEncoding:  strings.NewReplacer(strings.Split(replacements, "")...),
		replacerBeforeDecoding: strings.NewReplacer(strings.Split(util.ReverseString(replacements), "")...),
	}
}
