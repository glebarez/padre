package util

import "strings"

// ReverseString returns reverse of a string (does not support runes)
func ReverseString(in string) string {
	out := strings.Builder{}
	for i := len(in) - 1; i >= 0; i-- {
		out.WriteByte(in[i])
	}
	return out.String()
}
