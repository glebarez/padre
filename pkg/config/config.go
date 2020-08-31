package config

import (
	"github.com/glebarez/padre/pkg/encoding"
	"github.com/glebarez/padre/pkg/http"
)

// Config - all the settings
type Config struct {
	BlockLen                *int
	Parallel                *int
	URL                     *string
	Encoder                 encoding.EncoderDecoder
	PaddingErrorPattern     *string
	PaddingErrorFingerprint *http.ResponseFingerprint
	ProxyURL                *string
	POSTdata                *string
	ContentType             *string
	Cookies                 []*http.Cookie
	TermWidth               int
	EncryptMode             *bool
}
