package main

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
)

var hacky = true

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {

		if hacky {
			b[i] = byte(rand.Intn(126-33) + 33) // byte from ASCII printable range
		} else {
			b[i] = '_'
		}
	}
	return string(b)
}

func decode(s string) ([]byte, error) {
	// s = strings.Replace(s, "~", "=", -1)
	// s = strings.Replace(s, "-", "+", -1)
	// s = strings.Replace(s, "!", "/", -1)

	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func encode(data []byte) string {
	s := base64.StdEncoding.EncodeToString(data)

	// s = strings.Replace(s, "=", "~", -1)
	// s = strings.Replace(s, "+", "-", -1)
	// s = strings.Replace(s, "/", "!", -1)
	return s
}

func escapeChar(char byte) string {
	str := fmt.Sprintf("%+q", string(char))
	str = str[1 : len(str)-1]
	return strings.Replace(str, `\"`, `"`, 1)
}
