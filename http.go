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

var client *http.Client
var headers http.Header

func init() {
	// create http client
	proxyURL, _ := url.Parse("http://localhost:8080")
	client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
	//client = &http.Client{}

	// headers
	headers = http.Header{"Connection": {"keep-alive"}}
}

func isPaddingError(cipher []byte, ctx *context.Context) (bool, error) {
	// encode the cipher
	cipherEncoded := encode(cipher)
	// url.QueryEscape(
	url, err := url.Parse(fmt.Sprintf(baseURL, cipherEncoded)))
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

	// report about made request
	chanReq <- 1

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
