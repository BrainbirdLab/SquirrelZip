package lz77

import (
	"bytes"
	"fmt"
	"testing"
)

func TestLz77(t *testing.T) {
	CheckString(t, "abracadabra")
	CheckString(t, "banana")
	CheckString(t, "to be or not to be")
	CheckString(t, "the quick brown fox jumps over the lazy dog")
	CheckString(t, "this is a test string for lz77 compression and decompression")
	//CheckString(t, "lorem ipsum dolor sit amet, consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua ut enim ad minim veniam quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur excepteur sint occaecat cupidatat non proident sunt in culpa qui officia deserunt mollit anim id est laborum")
}

func CheckString(t *testing.T, input string) {
	inputBuffer := bytes.NewBuffer([]byte(input))
	outputBuffer := bytes.NewBuffer([]byte{})
	// Compress the input
	err := compressLZ77(inputBuffer, outputBuffer)
	if err != nil {
		t.Fatal(err)
	}

	// Decompress the output
	decompressedBuffer := bytes.NewBuffer([]byte{})
	err = decompressLZ77(outputBuffer, decompressedBuffer)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("input: %s\n", input)
	fmt.Printf("compressed: %s\n", outputBuffer.String())
	fmt.Printf("decompressed: %s\n", decompressedBuffer.String())
	//print size of input, compressed, and decompressed
	fmt.Printf("input size: %d\n", len(input))
	fmt.Printf("compressed size: %d\n", outputBuffer.Len())
	fmt.Printf("decompressed size: %d\n", decompressedBuffer.Len())

	// Check if the decompressed output is the same as the input
	if input != decompressedBuffer.String() {
		t.Fatalf("expected %s, got %s", input, decompressedBuffer.String())
	}

	fmt.Println()
}
