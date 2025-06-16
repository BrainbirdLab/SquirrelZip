package utils

import (
	"bytes"
	"io"
)

// ResizableBuffer provides a resettable buffer that can be written to and read from.
type ResizableBuffer struct {
	buf *bytes.Buffer
}

// NewResizableBuffer creates a new ResizableBuffer.
func NewResizableBuffer() *ResizableBuffer {
	return &ResizableBuffer{buf: new(bytes.Buffer)}
}

// Write appends the contents of p to the buffer, growing the buffer as needed.
func (rb *ResizableBuffer) Write(p []byte) (n int, err error) {
	return rb.buf.Write(p)
}

// Len returns the number of bytes currently in the buffer.
func (rb *ResizableBuffer) Len() int {
	return rb.buf.Len()
}

// Bytes returns a slice of length b.Len() holding the unread portion of the buffer.
func (rb *ResizableBuffer) Bytes() []byte {
	return rb.buf.Bytes()
}

// Reader returns a new io.Reader that reads from the underlying buffer.
func (rb *ResizableBuffer) Reader() io.Reader {
	return bytes.NewReader(rb.buf.Bytes())
}

// Reset resets the buffer to be empty.
func (rb *ResizableBuffer) Reset() {
	rb.buf.Reset()
}
