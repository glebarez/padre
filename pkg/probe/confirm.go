package probe

import (
	"context"

	"github.com/glebarez/padre/pkg/client"
	"github.com/glebarez/padre/pkg/util"
)

// confirms existence of padding oracle
// returns true if confirmed, false otherwise
func ConfirmPaddingOracle(c *client.Client, matcher PaddingErrorMatcher, blockLen int) (bool, error) {
	// create random block of ciphertext (IV prepended)
	cipher := util.RandomSlice(blockLen * 2)

	// test last byte of IV
	pos := blockLen - 1

	// channel to soak results
	chanResult := make(chan *client.ProbeResult, 256)

	// fingerprint probes
	go c.SendProbes(context.Background(), cipher, pos, chanResult)

	// send probes
	c.SendProbes(context.Background(), cipher, pos, chanResult)

	// count padding errors
	count := 0
	for result := range chanResult {
		if result.Err != nil {
			return false, result.Err
		}
		isErr, err := matcher.IsPaddingError(result.Response)
		if err != nil {
			return false, err
		}

		if isErr {
			count++
		}
	}

	// padding oracle must produce exactly 254 or 255 errors
	return count == 254 || count == 255, nil
}
