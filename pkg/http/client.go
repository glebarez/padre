package http

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/glebarez/padre/pkg/config"
	"github.com/glebarez/padre/pkg/encoding"
)

// Client - API to perform HTTP Requests to a remote server.
// Very specific for padre in that it sends queries to a specific URL
// that carries out the decryption and can spill padding oracle
type Client struct {
	// underlying net/http client
	client *http.Client

	// target URL that carries out the decryption and can spill padding oracle
	url string

	// pattern for HTTP POST data
	POSTdata string

	// HTTP concurrency (maximum number of simultaneous connections)
	concurrency int

	// encoder that is used to transform binary cipher into plaintext
	// this should comply with what remote server uses (e.g. Base64, Hex, etc)
	cipherEncoder encoding.EncoderDecoder
}

// Response - HTTP Response data
type Response struct {
	StatusCode int
	Data       []byte
}

// NewClient - Client Factory
func NewClient(config *config.Config) (*Client, error) {

	// parse proxy URL
	var (
		proxyURL *url.URL
		err      error
	)

	if proxy == "" {
		proxyURL = nil
	} else {
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			return nil, err
		}
	}

	// create net/http client
	client := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost: concurrency,
			Proxy:           http.ProxyURL(proxyURL),
		},
	}

	// return new client
	return &Client{
		client:      client,
		concurrency: concurrency,
	}, nil
}

// replace all occurrences of $ placeholder in a string, url-encoded if desired
func replacePlaceholder(s string, replacement string) string {
	replacement = url.QueryEscape(replacement)
	return strings.Replace(s, "$", replacement, -1)
}

// send HTTP request with cipher, encoded according to config
// returns response, response body, error
func (c *Client) doRequest(ctx context.Context, cipher []byte) (*Response, error) {
	// encode the cipher
	cipherEncoded := c.cipherEncoder.EncodeToString(cipher)

	// build URL
	url, err := url.Parse(replacePlaceholder(c.url, cipherEncoded))
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
		data := replacePlaceholder(c.POSTdata, cipherEncoded)
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
