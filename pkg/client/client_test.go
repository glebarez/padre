package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/glebarez/padre/pkg/encoder"
	"github.com/glebarez/padre/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestClient_DoRequest(t *testing.T) {
	// channel to propagate requests
	requestChan := make(chan *http.Request, 1)

	// special handler for propagating the channel
	handler := func(w http.ResponseWriter, r *http.Request) {
		// propagate received request
		requestChan <- r

		// copy request body into the response
		_, err := io.Copy(w, r.Body)
		assert.NoError(t, err)
	}

	// new test server
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// mock http client
	requestEventChan := make(chan byte, 1)

	// chose encoder
	encoder := encoder.NewB64encoder("")

	// create test client
	testURI := "/?data=$"

	client := &Client{
		HTTPclient:        ts.Client(),
		URL:               ts.URL + testURI,
		POSTdata:          "data=$",
		Cookies:           []*http.Cookie{{Name: "key", Value: "$"}},
		CipherPlaceholder: "$",
		Encoder:           encoder,
		Concurrency:       1,
		ContentType:       "cont/type",
		RequestEventChan:  requestEventChan,
	}

	// total requests to be sent
	totalRequestCount := 100

	// counter for received requests
	totalRequestsReceived := 0

	// asserts start here
	assert := assert.New(t)

	// send some requests with random data
	for i := 0; i < totalRequestCount; i++ {
		// generate random chunk
		data := util.RandomSlice(13)
		dataEncoded := encoder.EncodeToString(data)

		// send
		response, err := client.DoRequest(context.Background(), data)
		assert.NoError(err)

		// retrieve request event
		totalRequestsReceived += int(<-requestEventChan)

		// retrieve request that was sent to mocked http client
		request := <-requestChan

		// check URL formed properly
		assert.Equal(replacePlaceholder(testURI, "$", dataEncoded), request.RequestURI)

		// check Body formed properly
		assert.Equal(replacePlaceholder(client.POSTdata, "$", dataEncoded), string(response.Body))

		// check Cookie formed properly
		cookie, err := request.Cookie("key")
		assert.NoError(err)
		assert.Equal(url.QueryEscape(dataEncoded), cookie.Value)

		// check content type
		assert.Equal(request.Header.Get("Content-Type"), "cont/type")

	}

	// check total requests reported
	assert.Equal(totalRequestCount, totalRequestsReceived)
}

func TestClient_BrokenURL(t *testing.T) {
	client := &Client{URL: " http://foo.com", Encoder: encoder.NewB64encoder("")}
	_, err := client.DoRequest(context.Background(), []byte{})
	assert.Error(t, err)
}

func TestClient_NotRespondingServer(t *testing.T) {
	client := &Client{
		HTTPclient: http.DefaultClient,
		URL:        "http://localhost:1",
		Encoder:    encoder.NewB64encoder(""),
	}
	_, err := client.DoRequest(context.Background(), []byte{})
	assert.Error(t, err)
}
