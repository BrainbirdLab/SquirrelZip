package lz77

import (
	"bytes"
	"testing"
)

func TestCompressDecompress(t *testing.T) {
	input := []byte("this is a simple example of lz77 compression algorithm. this is a simple example of lz77 compression algorithm.")

	compressed, err := Compress(input)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(input, decompressed) {
		t.Errorf("Decompressed data does not match original.\nGot: %s,\nWant: %s", decompressed, input)
	}
}

func TestCompressEmptyInput(t *testing.T) {
	input := []byte{}

	compressed, err := Compress(input)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(input, decompressed) {
		t.Errorf("Decompressed data does not match original. Got: %s, Want: %s", decompressed, input)
	}
}

func TestCompressSingleCharacter(t *testing.T) {
	input := []byte("a")

	compressed, err := Compress(input)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(input, decompressed) {
		t.Errorf("Decompressed data does not match original. Got: %s, Want: %s", decompressed, input)
	}
}

func TestCompressNoRepetitions(t *testing.T) {
	input := []byte("abcdefghijklmnopqrstuvwxyz")

	compressed, err := Compress(input)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(input, decompressed) {
		t.Errorf("Decompressed data does not match original. Got: %s, Want: %s", decompressed, input)
	}
}
