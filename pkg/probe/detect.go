package probe

import (
	"context"
	"fmt"

	"github.com/glebarez/padre/pkg/client"
	"github.com/glebarez/padre/pkg/util"
)

// attempts to auto-detect padding oracle fingerprint
func detectPaddingErrorFingerprint(client *client.Client, blockLen int) (*paddingErrorTester, error) {
	// create random block of ciphertext (IV prepended)
	cipher := util.RandomSlice(blockLen * 2)

	// test last byte of IV
	pos := blockLen - 1

	// channel to soak results
	chanResult := make(chan *probeResult, 256)

	// fingerprint probes
	go sendProbes(context.Background(), client, cipher, pos, chanResult)

	// collect counts of fingerprints
	fpMap := map[ResponseFingerprint]int{}
	for result := range chanResult {
		if result.err != nil {
			// error during probes
			return nil, result.err
		}

		fp, err := GetResponseFingerprint(result.response)
		if err != nil {
			// error during fingerprinting
			return nil, result.err
		}

		fpMap[*fp]++
	}

	// padding oracle must respond with 254 or 255 identical fingerprints
	for fp, count := range fpMap {
		if count == 254 || count == 255 {
			return &paddingErrorTester{
				fingerprints: []*ResponseFingerprint{&fp},
			}, nil
		}
	}

	return nil, fmt.Errorf("Could not automaticaly detect padding error fingerprint")
}
