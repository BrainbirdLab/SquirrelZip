package compressor

import (
	"bytes"
	"encoding/gob"
	"os"
)

// WriteCompressedFile writes the compressed data, codes, and encryption flag to a file
func WriteCompressedFile(filename string, compressed []byte, codes map[rune]string, encrypted bool) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	// Write encryption flag
	if err := enc.Encode(encrypted); err != nil {
		return err
	}

	// Write codes
	if err := enc.Encode(codes); err != nil {
		return err
	}

	// Write compressed data
	if _, err := buf.Write(compressed); err != nil {
		return err
	}

	return os.WriteFile(filename, buf.Bytes(), 0644)
}

// ReadCompressedFile reads the compressed data and codes from a file and checks the encryption flag
func ReadCompressedFile(filename string) ([]byte, map[rune]string, bool, error) {
	// Read entire file content
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, false, err
	}

	// Initialize a buffer with file content
	buf := bytes.NewBuffer(data)

	// Create a decoder for the buffer
	dec := gob.NewDecoder(buf)

	// Read encryption flag
	var encrypted bool
	if err := dec.Decode(&encrypted); err != nil {
		return nil, nil, false, err
	}

	// Read codes
	var codes map[rune]string
	if err := dec.Decode(&codes); err != nil {
		return nil, nil, false, err
	}

    // Return the remaining data as compressed data
    return buf.Bytes(), codes, encrypted, nil
}
