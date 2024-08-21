package lz77

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	windowSize = 4096  // Larger sliding window
	maxMatch   = 258   // Maximum match length
	minMatch   = 2     // Minimum match length
)

type LZ77Compressor struct {
	window []byte
}

func NewLZ77Compressor() *LZ77Compressor {
	return &LZ77Compressor{
		window: make([]byte, 0, windowSize),
	}
}

func (c *LZ77Compressor) CompressData(data []byte) ([]byte, error) {
	var compressed bytes.Buffer
	windowStart := 0

	for i := 0; i < len(data); {
		matchOffset, matchLength := c.findLongestMatch(data, windowStart, i)
		
		if matchLength >= minMatch {
			// Write flag to indicate a match
			compressed.WriteByte(1)

			// Write offset and length
			offset := uint16(i - matchOffset)
			binary.Write(&compressed, binary.LittleEndian, offset)
			compressed.WriteByte(byte(matchLength))

			i += matchLength
		} else {
			// Write flag to indicate a literal byte
			compressed.WriteByte(0)
			compressed.WriteByte(data[i])
			i++
		}

		// Update the sliding window
		if len(c.window) >= windowSize {
			windowStart++
		}
		c.window = append(c.window, data[i-1])
		if len(c.window) > windowSize {
			c.window = c.window[1:]
		}
	}

	return compressed.Bytes(), nil
}

func (c *LZ77Compressor) findLongestMatch(data []byte, windowStart, current int) (int, int) {
	maxLen := min(maxMatch, len(data)-current)
	bestOffset := current
	bestLength := 0

	for offset := windowStart; offset < current; offset++ {
		length := 0
		for length < maxLen && data[offset+length] == data[current+length] {
			length++
		}
		if length > bestLength {
			bestLength = length
			bestOffset = offset
		}
	}

	return bestOffset, bestLength
}

func (c *LZ77Compressor) DecompressData(compressed []byte) ([]byte, error) {
	var uncompressed bytes.Buffer
	reader := bytes.NewReader(compressed)

	for reader.Len() > 0 {
		flag, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}

		if flag == 0 { // Literal byte
			if err := c.decompressLiteralByte(&uncompressed, reader); err != nil {
				return nil, err
			}
		} else if flag == 1 { // (offset, length) pair
			if err := c.decompressOffsetLength(&uncompressed, reader); err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("unknown flag")
		}
	}

	return uncompressed.Bytes(), nil
}

func (c *LZ77Compressor) decompressLiteralByte(uncompressed *bytes.Buffer, reader *bytes.Reader) error {
	literal, err := reader.ReadByte()
	if err != nil {
		return err
	}
	uncompressed.WriteByte(literal)
	return nil
}

func (c *LZ77Compressor) decompressOffsetLength(uncompressed *bytes.Buffer, reader *bytes.Reader) error {
	var offset uint16
	var length uint8
	if err := binary.Read(reader, binary.LittleEndian, &offset); err != nil {
		return err
	}
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return err
	}
	if int(offset) > uncompressed.Len() {
		return errors.New("invalid offset")
	}

	start := uncompressed.Len() - int(offset)
	for i := 0; i < int(length); i++ {
		uncompressed.WriteByte(uncompressed.Bytes()[start+i])
	}

	return nil
}