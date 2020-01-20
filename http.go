package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

var client = &http.Client{}
var headers = http.Header{"Connection": {"keep-alive"}}

// func elapsed(what string) func() {
// 	start := time.Now()
// 	return func() {
// 		log.Printf("%s took %v\n", what, time.Since(start))
// 	}
// }

func isPaddingError(cipher []byte, ctx *context.Context) (bool, error) {
	//time.Sleep(time.Second)
	//return false, nil

	// encode the cipher
	cipherEncoded := encode(cipher)

	url, err := url.Parse(fmt.Sprintf(baseURL, url.QueryEscape(cipherEncoded)))
	if err != nil {
		log.Fatal(err)
	}

	// create request
	req := &http.Request{
		URL: url,
		//Header: headers,
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
