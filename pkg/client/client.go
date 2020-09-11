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
	HTTPclient *http.Client

	// the following data will form the HTTP request payloads.
	// if placeholder is met among those data, it will be replaced
	// with encoded representation ciphertext
	URL      string
	POSTdata string
	Cookies  []*http.Cookie

	// placeholder to replace with encoded ciphertext
	CipherPlaceholder string

	// encoder that is used to transform binary ciphertext
	// into plaintext representation. this must comply with
	//  what remote server uses (e.g. Base64, Hex, etc)
	Encoder encoder.Encoder

	// HTTP concurrency (maximum number of simultaneous connections)
	Concurrency int

	// the content type of to be sent HTTP requests
	ContentType string

	// if this channel is not nil, it will be provided with byte value every time
	// the new HTTP request is made, so that RPS stats can be collected from
	// outside parties
	RequestEventChan chan byte
}

// DoRequest - send HTTP request with cipher, encoded according to config
func (c *Client) DoRequest(ctx context.Context, cipher []byte) (*Response, error) {
	// encode the cipher
	cipherEncoded := c.Encoder.EncodeToString(cipher)

	// build URL
	url, err := url.Parse(replacePlaceholder(c.URL, c.CipherPlaceholder, cipherEncoded))
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
		// perform data for POST body
		req.Method = "POST"
		data := replacePlaceholder(c.POSTdata, c.CipherPlaceholder, cipherEncoded)
		req.Body = ioutil.NopCloser(strings.NewReader(data))

		// set content type
		req.Header["Content-Type"] = []string{c.ContentType}
	}

	// add cookies if any
	if c.Cookies != nil {
		for _, cookie := range c.Cookies {
			// add cookies
			req.AddCookie(&http.Cookie{
				Name:  cookie.Name,
				Value: replacePlaceholder(cookie.Value, c.CipherPlaceholder, cipherEncoded),
			})
		}
	}

	// add context if passed
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	// send request
	resp, err := c.HTTPclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// report about made request to status
	if c.RequestEventChan != nil {
		c.RequestEventChan <- 1
	}

	// read body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{StatusCode: resp.StatusCode, Body: body}, nil
}
