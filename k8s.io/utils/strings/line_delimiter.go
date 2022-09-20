package strings

import (
	"bytes"
	"io"
	"strings"
)

type LineDelimiter struct {
	output    io.Writer
	delimiter []byte
	buf       bytes.Buffer
}

func NewLineDelimiter(output io.Writer, delimiter string) *LineDelimiter {
	return &LineDelimiter{output: output, delimiter: []byte(delimiter)}
}

func (ld *LineDelimiter) Write(buf []byte) (n int, err error) {
	return ld.buf.Write(buf)
}

func (ld *LineDelimiter) Flush() (err error) {
	lines := strings.Split(ld.buf.String(), "\n")
	for _, line := range lines {
		if _, err = ld.output.Write(ld.delimiter); err != nil {
			return
		}
		if _, err = ld.output.Write([]byte(line)); err != nil {
			return
		}
		if _, err = ld.output.Write(ld.delimiter); err != nil {
			return
		}
		if _, err = ld.output.Write([]byte("\n")); err != nil {
			return
		}
	}
	return
}
