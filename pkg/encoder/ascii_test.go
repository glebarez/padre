package encoder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_asciiEncoder_EncodeToString(t *testing.T) {
	e := NewASCIIencoder()

	type args struct {
		input []byte
	}
	tests := []struct {
		name string
		e    Encoder
		args args
		want string
	}{
		{"empty", e, args{[]byte(``)}, ``},
		{"nonascii", e, args{[]byte{0, 1, 255}}, `\x00\x01\xff`},
		{"ascii", e, args{[]byte(`test`)}, `test`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := asciiEncoder{}
			if got := e.EncodeToString(tt.args.input); got != tt.want {
				t.Errorf("asciiEncoder.EncodeToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_asciiEncoder_DecodeString(t *testing.T) {
	e := &asciiEncoder{}

	decode := func() {
		e.DecodeString("")
	}

	require.Panicsf(t, decode, "", "")

}
