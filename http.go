package main

import (
	"context"
	"fmt"
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
	client = &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost: *config.parallel,
			Proxy:           http.ProxyURL(proxyURL),
		},
	}

	// headers
	headers = http.Header{"Connection": {"keep-alive"}}

	return nil
}

func isPaddingError(cipher []byte, ctx *context.Context) (bool, error) {
	// encode the cipher
	cipherEncoded := config.encoder.encode(cipher)

	// build URL
	url, err := url.Parse(fmt.Sprintf(strings.Replace(*config.URL, "$", `%s`, 1), url.QueryEscape(cipherEncoded)))
	if err != nil {
		return false, err
	}

	// create request
	req := &http.Request{
		URL:    url,
		Header: headers,
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

	// test for padding oracle error string
	matched, err := regexp.Match(*config.paddingError, body)
	if err != nil {
		return false, err
	}
	return matched, nil
}
