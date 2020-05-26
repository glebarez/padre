package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var client *http.Client
var headers http.Header

func initHTTP() error {
	// parse proxy URL
	var proxyURL *url.URL
	if *config.proxyURL == "" {
		proxyURL = nil
	} else {
		var err error
		proxyURL, err = url.Parse(*config.proxyURL)
		if err != nil {
			return err
		}
	}

	// http client
	// TODO: more tweaking
	client = &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost: *config.parallel,
			Proxy:           http.ProxyURL(proxyURL),
		},
	}

	// headers map init
	headers = http.Header{}

	return nil
}

// replace all occurrences of $ placeholder in a string, url-encoded if desired
func replacePlaceholder(s string, replacement string) string {
	replacement = url.QueryEscape(replacement)
	return strings.Replace(s, "$", replacement, -1)
}

// send HTTP request with cipher, encoded according to config
// returns response, response body, error
func doRequest(ctx context.Context, cipher []byte) (*http.Response, []byte, error) {
	// encode the cipher
	cipherEncoded := config.encoder.EncodeToString(cipher)

	// build URL
	url, err := url.Parse(replacePlaceholder(*config.URL, cipherEncoded))
	if err != nil {
		return nil, nil, err
	}

	// create request
	req := &http.Request{
		URL:    url,
		Header: headers.Clone(),
	}

	// upgrade to POST if data is provided
	if *config.POSTdata != "" {
		req.Method = "POST"
		data := replacePlaceholder(*config.POSTdata, cipherEncoded)
		req.Body = ioutil.NopCloser(strings.NewReader(data))

		/* clone header before changing, so that:
		1. we don't mess the original template header variable
		2. to make it concurrency-save, otherwise expect panic */
		req.Header["Content-Type"] = []string{*config.contentType}
	}

	// add cookies if any
	if config.cookies != nil {
		for _, c := range config.cookies {
			// copy template cookie instance, so that we're concurrent-safe
			cookieCopy := &http.Cookie{Name: c.Name, Value: c.Value}
			cookieCopy.Value = replacePlaceholder(cookieCopy.Value, cipherEncoded)

			// add cookie to the requests
			req.AddCookie(cookieCopy)
		}
	}

	// add context if passed
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	// send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// report about made request to status
	reportHTTPRequest()

	// read body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, body, nil
}

/* function type that will process probe result, used in sendProbes() */
type probeFunc func(*http.Response, []byte) (interface{}, error)

type probeResult struct {
	probe  byte
	result interface{} // returned from probeFunc
	err    error       // returned from probeFunc
}

/* given a chunk of bytes, place every possible byte-value at specified position.
these probes are sent concurrently over HTTP, the responses will be processed by specified probeFunc
the results of probeFunc will be written into output channel
the channel is returned to the caller, to read from it*/
func sendProbes(ctx context.Context, chunk []byte, pos int, probeFunc probeFunc) chan *probeResult {
	/* how many unique probes will be run.
	equals to 2**8, since we're testing every possible value of a byte */
	const probeCount = 256

	/* communication channels, buffered for:
	- performance
	- to avoid goroutine leak due to deadlocks in edge cases*/
	chanIn := make(chan byte, probeCount)
	chanResult := make(chan *probeResult, probeCount)

	/* run workers */
	wg := sync.WaitGroup{}
	for i := 0; i < *config.parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// copy chunk to produce local concurrent-safe copy
			chunkCopy := sliceCopy(chunk)

			// do the work
			for {
				select {
				case <-ctx.Done():
					// early exit if context is cancelled
					return
				case b, ok := <-chanIn:
					// exit when input channel exhausted
					if !ok {
						return
					}

					// modify byte at given position
					chunkCopy[pos] = b

					// make HTTP request
					resp, body, err := doRequest(ctx, chunkCopy)
					if ctx.Err() == context.Canceled {
						return
					}

					// error during HTTP request
					if err != nil {
						chanResult <- &probeResult{
							probe: b,
							err:   err,
						}
						continue
					}

					// process probe result
					result, err := probeFunc(resp, body)
					chanResult <- &probeResult{
						probe:  b,
						result: result,
						err:    err,
					}
				}
			}
		}()
	}

	/* close output channel when workers are done */
	go func() {
		wg.Wait()
		close(chanResult)
	}()

	/* input generator: every possible byte value */
	go func() {
		for i := 0; i <= 0xff; i++ {
			chanIn <- byte(i)
		}
		close(chanIn)
	}()

	/* deliver output channel to caller */
	return chanResult
}
