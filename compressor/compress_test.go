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
	_, fileMeta, err := Compress(fileNameStrs, outputDir, "", algo)
	if err != nil {
		t.Fatalf("failed to compress files: %v", err)
	}

	fileMeta.PrintFileInfo()
	fileMeta.PrintCompressionRatio()
}

func TestDecompress(t *testing.T) {
	InitDecompress(t)
}

func InitDecompress(t *testing.T) {
	
	compressedFilePath := "test_files/compress_output/compressed.sq"
	outputDir := "test_files/decompress_output"

	filePaths, err := Decompress(compressedFilePath, outputDir, "")
	if err != nil {
		t.Fatalf("failed to decompress file: %v", err)
	}
	
	for _, filePath := range filePaths {
		fmt.Printf("Decompressed file: %s\n", filePath)
	}
}

func TestCompressFile(t *testing.T) {
	CompressFile("test_files/compress_output/compressed.sq", "test_files/compress_output_2", "huffman", t)
}

func CompressFile(filePath, outputDir, algo string, t *testing.T) {
	_, fileMeta, err := Compress([]string{filePath}, outputDir, "", algo)
	if err != nil {
		t.Fatalf("failed to compress file: %v", err)
	}

	fileMeta.PrintFileInfo()
	fileMeta.PrintCompressionRatio()
}