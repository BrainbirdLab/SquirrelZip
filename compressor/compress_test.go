package compressor

import (
	"fmt"
	"os"
	"testing"

	//"file-compressor/utils"
)

func TestCompress(t *testing.T) {
	Init("huffman", t)
	//Init("arithmetic", t)
}

func Init(algo string, t *testing.T) {

	testFilesDir := "test_files"
	testFiles, err := os.ReadDir(testFilesDir)
	if err != nil {
		t.Fatalf("failed to read test files directory: %v", err)
	}

	fileNameStrs := make([]string, 0)
	for _, file := range testFiles {
		fileNameStrs = append(fileNameStrs, fmt.Sprintf("%s/%s", testFilesDir, file.Name()))
		fmt.Printf("Found file: %s\n", file.Name())
	}

	outputDir := "output"
	outputPath, size, err := Compress(fileNameStrs, outputDir, "", algo)
	if err != nil {
		t.Fatalf("failed to compress files: %v", err)
	}
	fmt.Printf("Compressed file: %s\nSize: %d\n", outputPath, size)

	// Decompress
	decompressedPath, err := Decompress(outputPath, outputDir, algo)
}