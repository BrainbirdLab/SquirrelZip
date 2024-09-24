package lampelziv

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// DecompressData decompresses the given compressed data using a basic Lempel-Ziv algorithm.
func DecompressData(compressed []byte) ([]byte, error) {
	uncompressed := bytes.Buffer{}
	reader := bytes.NewReader(compressed)

	for reader.Len() > 0 {
		flag, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("failed to read flag byte: %v", err)
		}

		switch flag {
		case 0:
			err := handleLiteral(reader, &uncompressed)
			if err != nil {
				return nil, err
			}
		case 1:
			err := handleMatch(reader, &uncompressed)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown flag in compressed data: %d", flag)
		}
	}

	return uncompressed.Bytes(), nil
}

func handleLiteral(reader *bytes.Reader, uncompressed *bytes.Buffer) error {
	literal, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read literal byte: %v", err)
	}
	uncompressed.WriteByte(literal)
	return nil
}

func handleMatch(reader *bytes.Reader, uncompressed *bytes.Buffer) error {
	var offset uint8
	var length uint8

	if err := binary.Read(reader, binary.LittleEndian, &offset); err != nil {
		return fmt.Errorf("failed to read offset: %v", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return fmt.Errorf("failed to read length: %v", err)
	}

	// Validate offset and length
	start := uncompressed.Len() - int(offset)
	if start < 0 || start+int(length) > uncompressed.Len() {
		return fmt.Errorf("invalid length for substring: start=%d, length=%d, buffer length=%d", start, length, uncompressed.Len())
	}

	fmt.Printf("Offset: %d, Length: %d\n", offset, length) // Debugging: print offset and length

	// Extract and append the match
	substr := uncompressed.Bytes()[start : start+int(length)]
	uncompressed.Write(substr)

	return nil
}