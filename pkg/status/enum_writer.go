package status

import (
	"io"
)

var (
	lf    = byte(0xa)
	space = []byte(" ")
)

type PrefixedWriter struct {
	prefix     []byte
	indent     []byte
	writer     io.Writer
	wasWritten bool
}

func (pw *PrefixedWriter) Write(p []byte) (n int, err error) {
	// the very first write is prefixed
	if !pw.wasWritten {
		pw.writer.Write(pw.prefix)
		pw.wasWritten = true
	}

	return pw.writer.Write(p)

}
