package client

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/glebarez/padre/pkg/encoder"
	"github.com/glebarez/padre/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestClient_DoRequest(t *testing.T) {
	// require start here
	require := require.New(t)

	// channel to propagate requests
	requestChan := make(chan *http.Request, 1)

	// special handler for propagating the channel
	handler := func(w http.ResponseWriter, r *http.Request) {
		// propagate received request
		requestChan <- r

		// copy request body into the response
		responseBody, err := ioutil.ReadAll(r.Body)
		require.NoError(err)

		// fill the response writer
		_, err = w.Write(responseBody)
		require.NoError(err)
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

	// send some requests with random data
	for i := 0; i < totalRequestCount; i++ {
		// generate random chunk
		data := util.RandomSlice(13)
		dataEncoded := encoder.EncodeToString(data)

		// send
		response, err := client.DoRequest(context.Background(), data)
		require.NoError(err)

		// retrieve request event
		totalRequestsReceived += int(<-requestEventChan)

		// retrieve request that was sent to mocked http client
		request := <-requestChan

		// check URL formed properly
		require.Equal(replacePlaceholder(testURI, "$", dataEncoded), request.RequestURI)

		// check Body formed properly
		require.Equal(replacePlaceholder(client.POSTdata, "$", dataEncoded), string(response.Body))

		// check Cookie formed properly
		cookie, err := request.Cookie("key")
		require.NoError(err)
		require.Equal(url.QueryEscape(dataEncoded), cookie.Value)

		// check content type
		require.Equal(request.Header.Get("Content-Type"), "cont/type")

	}

	// check total requests reported
	require.Equal(totalRequestCount, totalRequestsReceived)
}

func TestClient_BrokenURL(t *testing.T) {
	client := &Client{URL: " http://foo.com", Encoder: encoder.NewB64encoder("")}
	_, err := client.DoRequest(context.Background(), []byte{})
	require.Error(t, err)
}

func TestClient_NotRespondingServer(t *testing.T) {
	client := &Client{
		HTTPclient: http.DefaultClient,
		URL:        "http://localhost:1",
		Encoder:    encoder.NewB64encoder(""),
	}
	_, err := client.DoRequest(context.Background(), []byte{})
	require.Error(t, err)
}
