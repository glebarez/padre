package status

import (
	"bytes"
	"io"

	"github.com/glebarez/padre/pkg/color"
)

var (
	lf    = byte(0xa)
	space = []byte(" ")
)

type prefixedWriter struct {
	prefix     []byte
	indent     []byte
	writer     io.Writer
	wasWritten bool
}

func (pw *prefixedWriter) Write(p []byte) (n int, err error) {
	// the very first write is prefixed
	if !pw.wasWritten {
		pw.writer.Write(pw.prefix)
		pw.wasWritten = true
	}

	return pw.writer.Write(p)
}

func NewPrefixedWriter(prefix string, writer io.Writer) io.Writer {
	return &prefixedWriter{
		prefix: []byte(prefix),
		indent: bytes.Repeat(space, len(color.StripColor(prefix))),
		writer: writer,
	}
}
