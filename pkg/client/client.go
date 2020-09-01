package client

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/glebarez/padre/pkg/encoder"
)

// Client - API to perform HTTP Requests to a remote server.
// Very specific to padre, in that it sends queries to a specific URL
// that carries out the decryption and can spill padding oracle
type Client struct {
	// underlying net/http client
	client *http.Client

	// the following data will form the HTTP request payloads.
	// if placeholder is met among those data, it will be replaced
	// with encoded representation ciphertext
	url      string
	POSTdata string
	cookies  []*http.Cookie

	// placeholder to replace with encoded ciphertext
	cihperPlaceholder string

	// encoder that is used to transform binary ciphertext
	// into plaintext representation. this must comply with
	//  what remote server uses (e.g. Base64, Hex, etc)
	encoder encoder.Encoder

	// HTTP concurrency (maximum number of simultaneous connections)
	Concurrency int

	// the content type of to be sent HTTP requests
	contentType string

	// stats collector: this channel (if set) will be provided with
	// byte value (0x00) every time the new HTTP request is made.
	// (external caller can use it to calculate RPS)
	newRequestEventHandler *func()
}

// type NewClientRequest struct {
// 	ProxyURL string
// 	TargetURL string
// 	POSTdata string
// 	Cookies string
// 	ContentType string
// 	CipherPlaceholder string
// 	Encoder encoder.Encoder
// }

// // NewClient - Client Factory
// func NewClient(proxy string, concurrency int) (*Client, error) {
// 	// create net/http client
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			MaxConnsPerHost: concurrency,
// 			Proxy:           http.ProxyURL(proxyURL),
// 		},
// 	}

// 	// return new client
// 	return &Client{
// 		client:      client,
// 		concurrency: concurrency,
// 	}, nil
// }

// DoRequest - send HTTP request with cipher, encoded according to config
func (c *Client) DoRequest(ctx context.Context, cipher []byte) (*Response, error) {
	// encode the cipher
	cipherEncoded := c.encoder.EncodeToString(cipher)

	// build URL
	url, err := url.Parse(replacePlaceholder(c.url, cipherEncoded))
	if err != nil {
		return nil, err
	}

	// create request
	req := &http.Request{
		URL:    url,
		Header: http.Header{},
	}

	// upgrade to POST if data is provided
	if c.POSTdata != "" {
		req.Method = "POST"
		data := replacePlaceholder(c.POSTdata, cipherEncoded)
		req.Body = ioutil.NopCloser(strings.NewReader(data))

		/* clone header before changing, so that:
		1. we don't mess the original template header variable
		2. to make it concurrency-save, otherwise expect panic */
		req.Header["Content-Type"] = []string{c.contentType}
	}

	// add cookies if any
	if c.cookies != nil {
		for _, c := range c.cookies {
			// add cookies
			req.AddCookie(&http.Cookie{
				Name:  c.Name,
				Value: replacePlaceholder(c.Value, cipherEncoded),
			})
		}
	}

	// add context if passed
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	// send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// report about made request to status
	if c.newRequestEventHandler != nil {
		(*c.newRequestEventHandler)()
	}

	// read body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{StatusCode: resp.StatusCode, Body: body}, nil
}
