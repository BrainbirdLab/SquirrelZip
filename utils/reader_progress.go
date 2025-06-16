package utils

import (
	"io"
)

// ProgressTrackingReader is an io.Reader that wraps an existing reader
// to track the number of bytes read and call a progress callback.
type ProgressTrackingReader struct {
	reader     io.Reader
	totalSize  int64
	readSoFar  int64
	onProgress func(readBytes int64, totalBytes int64)
}

// NewProgressReader creates a new ProgressTrackingReader.
func NewProgressReader(r io.Reader, totalSize int64, onProgress func(readBytes int64, totalBytes int64)) io.Reader {
	return &ProgressTrackingReader{
		reader:     r,
		totalSize:  totalSize,
		onProgress: onProgress,
	}
}

// Read reads up to len(p) bytes into p.
func (r *ProgressTrackingReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.readSoFar += int64(n)
	if r.onProgress != nil {
		r.onProgress(r.readSoFar, r.totalSize)
	}
	return n, err
}
