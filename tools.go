package main

import (
	"encoding/base64"
	"strings"
)

func decode(s string) ([]byte, error) {
	s = strings.Replace(s, "~", "=", -1)
	s = strings.Replace(s, "-", "+", -1)
	s = strings.Replace(s, "!", "/", -1)

	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func encode(data []byte) string {
	s := base64.StdEncoding.EncodeToString(data)

	s = strings.Replace(s, "=", "~", -1)
	s = strings.Replace(s, "+", "-", -1)
	s = strings.Replace(s, "/", "!", -1)
	return s
}
