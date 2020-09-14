package encoder

// Encoder - performs encoding/decoding
type Encoder interface {
	EncodeToString([]byte) string
	DecodeString(string) ([]byte, error)
}

// DecodeError ...
type DecodeError string

func (e DecodeError) Error() string { return string(e) }
