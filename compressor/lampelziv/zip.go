package lampelziv

import (
	"bytes"
	"encoding/binary"
)

// CompressData compresses the input data using a basic Lempel-Ziv algorithm.
func CompressData(content []byte) ([]byte, error) {
	var compressed bytes.Buffer
	windowSize := 1024
	currentPos := 0
	bufferSize := len(content)

	for currentPos < bufferSize {
		bestMatchLength, bestMatchOffset := findBestMatch(content, currentPos, windowSize, bufferSize)

		if bestMatchLength >= 3 { // Minimum match length threshold
			// Encode the match with a flag of 1
			compressed.WriteByte(1) // Match flag
			binary.Write(&compressed, binary.LittleEndian, uint8(bestMatchOffset))
			binary.Write(&compressed, binary.LittleEndian, uint8(bestMatchLength))
			currentPos += bestMatchLength
		} else {
			// Encode literal with a flag of 0
			compressed.WriteByte(0) // Literal flag
			compressed.WriteByte(content[currentPos])
			currentPos++
		}
	}

	return compressed.Bytes(), nil
}

func findBestMatch(content []byte, currentPos, windowSize, bufferSize int) (int, int) {
	bestMatchLength := 0
	bestMatchOffset := 0

	// Search for the best match within the sliding window
	for offset := max(0, currentPos-windowSize); offset < currentPos; offset++ {
		length := 0
		for length < bufferSize-currentPos &&
			content[offset+length] == content[currentPos+length] {
			length++
			if offset+length >= currentPos {
				break
			}
		}

		if length > bestMatchLength {
			bestMatchLength = length
			bestMatchOffset = currentPos - offset
		}
	}

	return bestMatchLength, bestMatchOffset
}
