package lz77

import (
	"encoding/binary"
	"io"
)

type Token struct {
	Offset int16
	Length int16
	Char   byte
}

const windowSize = 20
const chunkSize = 32

// LZ77 compression function using io.Reader and io.Writer
func compressLZ77(r io.Reader, w io.Writer) error {
	buffer := make([]byte, chunkSize)
	var slidingWindow []byte

	for {
		// Read the next chunk
		n, err := r.Read(buffer)
		if n == 0 {
			break
		}

		if err != nil && err != io.EOF {
			return err
		}

		inputBytes := buffer[:n]

		// Process each byte in the chunk
		for i := 0; i < len(inputBytes); {
			start := i - len(slidingWindow)
			if start < 0 {
				start = 0
			}
			window := slidingWindow[start:]

			// Find the longest match in the window
			var longestOffset, longestLength int16
			for j := 0; j < len(window); j++ {
				length := 0
				for length < len(window)-j && i+length < len(inputBytes) && window[j+length] == inputBytes[i+length] {
					length++
				}
				if length > int(longestLength) {
					longestOffset = int16(len(window) - j)
					longestLength = int16(length)
				}
			}

			nextChar := byte(0)
			if i+int(longestLength) < len(inputBytes) {
				nextChar = inputBytes[i+int(longestLength)]
			}

			// Write the token to io.Writer
			binary.Write(w, binary.LittleEndian, Token{Offset: longestOffset, Length: longestLength, Char: nextChar})

			// Add processed data to sliding window
			slidingWindow = append(slidingWindow, inputBytes[i:i+int(longestLength)+1]...)
			if len(slidingWindow) > windowSize {
				slidingWindow = slidingWindow[len(slidingWindow)-windowSize:]
			}

			i += int(longestLength) + 1
		}
	}

	return nil
}

// LZ77 decompression function using io.Reader and io.Writer
func decompressLZ77(r io.Reader, w io.Writer) error {
	var slidingWindow []byte
	var token Token

	for {
		// Read the token from io.Reader
		err := binary.Read(r, binary.LittleEndian, &token)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// Get the match from the sliding window
		start := len(slidingWindow) - int(token.Offset)
		for i := 0; i < int(token.Length); i++ {
			w.Write([]byte{slidingWindow[start+i]})
			slidingWindow = append(slidingWindow, slidingWindow[start+i])
		}

		// Add the next character
		if token.Char != 0 {
			w.Write([]byte{token.Char})
			slidingWindow = append(slidingWindow, token.Char)
		}

		// Maintain sliding window size
		if len(slidingWindow) > windowSize {
			slidingWindow = slidingWindow[len(slidingWindow)-windowSize:]
		}
	}

	return nil
}
