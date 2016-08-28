package vim

import "io"

// NewReadWriter returns simple io.ReadWriter.
// bufio.ReadWriter has buffers and needs to be flushed., so we cannot use
// bufio.NewReadWriter() for Vim client which accept io.ReadWriter.
// ref: https://groups.google.com/forum/#!topic/golang-nuts/OJnnwlfsPCc
func NewReadWriter(r io.Reader, w io.Writer) io.ReadWriter {
	return &readWriter{r, w}
}

type readWriter struct {
	io.Reader
	io.Writer
}
