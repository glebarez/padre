package main

import (
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
		return 0, err
	}
	w, _ := termbox.Size()
	termbox.Close()
	return w, nil
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
