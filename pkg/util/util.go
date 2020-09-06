package util

import (
	"bytes"
	"container/ring"
	"strings"
)

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
