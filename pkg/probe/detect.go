package probe

import (
	"context"

	"github.com/glebarez/padre/pkg/client"
	"github.com/glebarez/padre/pkg/util"
)

// attempts to auto-detect padding oracle fingerprint
func detectPaddingErrorFingerprint(c *client.Client, blockLen int) (PaddingErrorMatcher, error) {
	// create random block of ciphertext (IV prepended)
	cipher := util.RandomSlice(blockLen * 2)

	// test last byte of IV
	pos := blockLen - 1

	// channel to soak results
	chanResult := make(chan *client.ProbeResult, 256)

	// fingerprint probes
	go c.SendProbes(context.Background(), cipher, pos, chanResult)

	// collect counts of fingerprints
	fpMap := map[ResponseFingerprint]int{}
	for result := range chanResult {
		if result.Err != nil {
			// error during probes
			return nil, result.Err
		}

		fp, err := GetResponseFingerprint(result.Response)
		if err != nil {
			// error during fingerprinting
			return nil, result.Err
		}

		fpMap[*fp]++
	}

	// padding oracle must respond with 254 or 255 identical fingerprints
	for fp, count := range fpMap {
		if count == 254 || count == 255 {
			return &matcherByFingerprint{
				fingerprints: []*ResponseFingerprint{&fp},
			}, nil
		}
	}

	return nil, nil
}
