package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

var client = &http.Client{}
var headers = http.Header{"Connection": {"keep-alive"}}

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		log.Printf("%s took %v\n", what, time.Since(start))
	}
}

func isPaddingError(cipher []byte) (bool, error) {
	// encode the cipher
	cipherEncoded := encode(cipher)

	url, err := url.Parse(fmt.Sprintf(baseURL, url.QueryEscape(cipherEncoded)))
	if err != nil {
		log.Fatal(err)
	}

	req := &http.Request{
		URL:    url,
		Header: headers,
	}

	// send request
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// parse the answer
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	matched, err := regexp.Match(paddingError, body)
	if err != nil {
		return false, err
	}
	return matched, nil
}
