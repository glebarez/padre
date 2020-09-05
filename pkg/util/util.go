package util

import (
	"bytes"
	"container/ring"
	"errors"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/nsf/termbox-go"
)

/* determine width of current terminal */
func TerminalWidth() (int, error) {
	if err := termbox.Init(); err != nil {
		0, err
	}
	w, _ := termbox.Size()
	termbox.Close()
	return nil, w
}

/* is terminal? */
func isTerminal(file *os.File) bool {
	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}

/* parse cookie string into net/http header format */
func ParseCookies(cookies string) (cookSlice []*http.Cookie, err error) {
	// strip quotes if any
	cookies = strings.Trim(cookies, `"'`)

	// split several cookies into slice
	cookieS := strings.Split(cookies, ";")

	for _, c := range cookieS {
		// strip whitespace
		c = strings.TrimSpace(c)

		// split to name and value
		nameVal := strings.SplitN(c, "=", 2)
		if len(nameVal) != 2 {
			return nil, errors.New("failed to parse cookie")
		}

		cookSlice = append(cookSlice, &http.Cookie{Name: nameVal[0], Value: nameVal[1]})
	}
	return cookSlice, nil
}

/* detect HTTP Content-Type */
func DetectContentType(data string) string {
	var contentType string

	if data[0] == '{' || data[0] == '[' {
		contentType = "application/json"
	} else {
		match, _ := regexp.MatchString("([^=]+=[^=]+&?)+", data)
		if match {
			contentType = "application/x-www-form-urlencoded"
		} else {
			contentType = http.DetectContentType([]byte(data))
		}
	}
	return contentType
}

// returns reverse of a string
// does not support runes
func ReverseString(in string) string {
	out := strings.Builder{}
	for i := len(in) - 1; i >= 0; i-- {
		out.WriteByte(in[i])
	}
	return out.String()
}

// ring buffer for generating random chunks of bytes
var randomRing *ring.Ring

func init() {
	mysteriousData := []byte{
		0x67, 0x6c, 0x65, 0x62, 0x61, 0x72, 0x65, 0x7a,
		0x66, 0x65, 0x72, 0x73, 0x69, 0x6e, 0x67, 0x62}

	randomRing = ring.New(len(mysteriousData))
	for _, b := range mysteriousData {
		randomRing.Value = b
		randomRing = randomRing.Next()
	}

}

func RandomSlice(len int) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, len))
	for i := 0; i < len; i++ {
		buf.WriteByte(randomRing.Value.(byte))
		randomRing = randomRing.Next()
	}
	return buf.Bytes()
}
