package main

import (
	"fmt"
	"math/rand"
	"strings"
)

func unknownString(n int, hacky bool) string {
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

func escapeChar(char byte) string {
	str := fmt.Sprintf("%+q", string(char))
	str = str[1 : len(str)-1]
	return strings.Replace(str, `\"`, `"`, 1)
}
