package compressor

import (
	"fmt"
	"os"
	"testing"
)

func TestCompress(t *testing.T) {
	Init("huffman", t)
	//Init("arithmetic", t)
}

func Init(algo string, t *testing.T) {
	testFilesDir := "test_files/input"
	testFiles, err := os.ReadDir(testFilesDir)
	if err != nil {
		t.Fatalf("failed to read test files directory: %v", err)
	}

	fileNameStrs := make([]string, 0)
	for _, file := range testFiles {
		fileNameStrs = append(fileNameStrs, fmt.Sprintf("%s/%s", testFilesDir, file.Name()))
	}

	outputDir := "test_files/compress_output"
	_, fileMeta, err := Compress(fileNameStrs, outputDir, algo)
	if err != nil {
		t.Fatalf("failed to compress files: %v", err)
	}

	fileMeta.PrintFileInfo()
	fileMeta.PrintCompressionRatio()
}