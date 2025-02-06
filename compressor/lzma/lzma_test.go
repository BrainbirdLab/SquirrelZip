package lzma

import (
	"bytes"
	"testing"
)

// testCase represents a single compression test case
type testCase struct {
	name  string
	input []byte
}

// testCompressDecompress is a helper function that runs the compression/decompression test
func testCompressDecompress(t *testing.T, tc testCase) {
	t.Helper()

	compressed, err := compressData(tc.input)
	if err != nil {
		t.Fatalf("%s: Compression failed: %v", tc.name, err)
	}

	decompressed, err := decompressData(compressed)
	if err != nil {
		t.Fatalf("%s: Decompression failed: %v", tc.name, err)
	}

	if !bytes.Equal(tc.input, decompressed) {
		t.Errorf("%s: Decompressed data does not match original.\nGot: %s\nWant: %s",
			tc.name, decompressed, tc.input)
	}
}

func TestCompression(t *testing.T) {
	tests := []testCase{
		{
			name: "Very Long Text",
			input: []byte("lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
				"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
				"ut enim ad minim veniam, quis nostrud exercitation ullamco laboris " +
				"nisi ut aliquip ex ea commodo consequat. duis aute irure dolor in " +
				"reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla " +
				"pariatur. excepteur sint occaecat cupidatat non proident, sunt in " +
				"culpa qui officia deserunt mollit anim id est laborum."),
		},
		{
			name: "Repeated Text",
			input: []byte("this is a simple example of lz77 compression algorithm. " +
				"this is a simple example of lz77 compression algorithm."),
		},
		{
			name:  "Empty Input",
			input: []byte{},
		},
		{
			name:  "Single Character",
			input: []byte("a"),
		},
		{
			name:  "No Repetitions",
			input: []byte("abcdefghijklmnopqrstuvwxyz"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testCompressDecompress(t, tc)
		})
	}
}
