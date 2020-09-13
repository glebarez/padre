package probe

import (
	"context"

	"github.com/glebarez/padre/pkg/client"
	"github.com/glebarez/padre/pkg/util"
)

// attempts to auto-detect padding oracle fingerprint
func DetectPaddingErrorFingerprint(c *client.Client, blockLen int) (PaddingErrorMatcher, error) {
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

	// padding oracles respond with predictable count of unique fingerprints
	// following factors must be considered:
	// a. some padding implmementations 'incorrect' padding from 'errornous' padding
	// (e.g. if you pad cipher with block length of 16 with values grater than 16)

	// padre considers following fingerprint counts as indication of padding error
	patterns := [][]int{
		{255, 1},
		{254, 2},
		{256 - blockLen, blockLen - 1, 1},
		{256 - blockLen, blockLen - 2, 2},
	}

	// check if any of count-patterns matches
patternLoop:
	for _, pat := range patterns {
		fingerprints := make([]ResponseFingerprint, 0)

		for fp, count := range fpMap {
			if inSlice(pat, count) {
				// do not include fingerprint of non-error response (last position in pattern)
				if count != pat[len(pat)-1] {
					fingerprints = append(fingerprints, fp)
				}
			} else {
				continue patternLoop
			}
		}

		// if we made it to here, we found a padding oracle
		// return the matcher
		return &matcherByFingerprint{
			fingerprints: fingerprints,
		}, nil
	}
	return nil, nil
}

func inSlice(slice []int, value int) bool {
	for _, i := range slice {
		if value == i {
			return true
		}
	}
	return false
}
