package encoder

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/glebarez/padre/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestReplacer(t *testing.T) {

	// test cases
	tests := []struct {
		name        string
		encoder     Encoder
		replFactory func(replacements string) Encoder
		replString  string
	}{
		{"b64", base64.StdEncoding, NewB64encoder, `=~/!+^`},
		{"lhex", &lhexEncoder{}, NewLHEXencoder, `0zfyeT`},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// generate random byte string
			byteData := util.RandomSlice(20)

			// encode with basic encoder
			encodedData := tt.encoder.EncodeToString(byteData)

			// replace characters
			encodedData = strings.NewReplacer(strings.Split(tt.replString, "")...).Replace(encodedData)

			// compare results
			replacer := tt.replFactory(tt.replString)
			require.Equal(t, replacer.EncodeToString(byteData), encodedData)

			// decode back and compare
			decoded, err := replacer.DecodeString(encodedData)
			require.NoError(t, err)
			require.Equal(t, decoded, byteData)

			// try decoding corrupted string
			_, err = replacer.DecodeString(string(encodedData[:len(encodedData)-1]))
			require.Error(t, err)
		})
	}
}
