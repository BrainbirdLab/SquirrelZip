package compressor

import (
	"fmt"
	"os"
	"testing"
)

func TestCompress(t *testing.T) {
	DecompressStart(Init("huffman", t), t)
}

func Init(algo string, t *testing.T) string {
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
	outputPath, fileMeta, err := Compress(fileNameStrs, outputDir, algo)
	if err != nil {
		t.Fatalf("failed to compress files: %v", err)
	}

	fileMeta.PrintFileInfo()
	fileMeta.PrintCompressionRatio()

	fmt.Println("Compression done: ", outputPath)

	return outputPath
}

func DecompressStart(compressedPath string, t *testing.T) {
	fmt.Printf("Decompressing file: %s\n", compressedPath)
	_, err := Decompress(compressedPath, "test_files/decompressed_output")
	if err != nil {
		t.Fatalf("failed to decompress files: %v", err)
	}

	fmt.Println("Decompression done")
	// delete compressed file and decompressed files
	err = os.RemoveAll("test_files/compress_output")
	if err != nil {
		t.Fatalf("failed to delete compressed file: %v", err)
	}
	err = os.RemoveAll("test_files/decompressed_output")
	if err != nil {
		t.Fatalf("failed to delete decompressed files: %v", err)
	}
}