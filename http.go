package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
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
			MaxConnsPerHost: *config.parallel * 2,
			Proxy:           http.ProxyURL(proxyURL),
		},
	}

	// headers
	headers = http.Header{"Connection": {"keep-alive"}}

	return nil
}

// replace cipher placeholder in a string with URL-escaped cipher
func replaceCipherPlaceholder(s string, cipherEncoded string) string {
	return strings.Replace(s, "$", url.QueryEscape(cipherEncoded), 1)
}

func isPaddingError(cipher []byte, ctx *context.Context) (bool, error) {
	// encode the cipher
	cipherEncoded := config.encoder.encode(cipher)

	// build URL
	url, err := url.Parse(replaceCipherPlaceholder(*config.URL, cipherEncoded))
	if err != nil {
		return false, err
	}

	// create request
	req := &http.Request{
		URL:    url,
		Header: headers,
	}

	// upgrade to POST if data is provided
	if *config.POSTdata != "" {
		req.Method = "POST"
		data := replaceCipherPlaceholder(*config.POSTdata, cipherEncoded)
		req.Body = ioutil.NopCloser(strings.NewReader(data))

		/* clone header before changing, so that:
		1. we don't mess the original template header variable
		2. to make it concurrency-save, otherwise expect panic */
		req.Header = req.Header.Clone()
		req.Header["Content-Type"] = []string{*config.contentType}
	}

	// add cookies if any
	if config.cookies != nil {
		for _, c := range config.cookies {
			// copy template cookie instance, so that we're concurrent-safe
			cookieCopy := &http.Cookie{Name: c.Name, Value: c.Value}
			cookieCopy.Value = replaceCipherPlaceholder(cookieCopy.Value, cipherEncoded)

			// add cookie to the requests
			req.AddCookie(cookieCopy)
		}
	}

	// add context if passed
	if ctx != nil {
		req = req.WithContext(*ctx)
	}

	// send request
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// report about made request
	currentStatus.reportHTTPRequest()

	// parse the answer
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// test for padding oracle error pattern
	matched, err := regexp.Match(*config.paddingError, body)
	if err != nil {
		return false, err
	}
	return matched, nil
}
