package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/glebarez/padre/pkg/encoder"
	"github.com/glebarez/padre/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestClient_SendProbes(t *testing.T) {
	reqBodyChan := make(chan []byte, 1)

	// special handler for propagating request body into channel
	handler := func(w http.ResponseWriter, r *http.Request) {
		// copy request body into the response
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		reqBodyChan <- body
		fmt.Fprintln(w, "grabbed")
	}

	// new test server
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// chose encoder
	encoder := encoder.NewB64encoder("")

	// create test client
	testURI := "/"

	client := &Client{
		HTTPclient:        ts.Client(),
		URL:               ts.URL + testURI,
		POSTdata:          "$",
		CipherPlaceholder: "$",
		Encoder:           encoder,
		Concurrency:       1,
	}

	// generate random chunk
	data := util.RandomSlice(20)

	// create channel for probe results
	chanProbeResult := make(chan *ProbeResult, 1)

	// test every position for a probe
	for pos := 0; pos < len(data); pos++ {
		// send probes
		go client.SendProbes(context.Background(), data, pos, chanProbeResult)

		// get probe result
		for probeResult := range chanProbeResult {
			assert.NoError(t, probeResult.Err)

			// derive expected probe data
			expectedProbe := copySlice(data)
			expectedProbe[pos] = probeResult.Byte

			// derive made probe data
			// get request body received by the test server
			requestBody, err := url.QueryUnescape(string(<-reqBodyChan))
			assert.NoError(t, err)

			madeProbe, err := encoder.DecodeString(requestBody)
			assert.NoError(t, err)

			// compare the two
			assert.Equal(t, expectedProbe, madeProbe)
		}
	}
}
