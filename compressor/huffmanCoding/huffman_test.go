package huffmanCoding

import (
	"file-compressor/utils"
	"fmt"
	"os"
	"testing"
)

func TestHuffman(t *testing.T) {

	testData := "Hello, World!"

	compressedData, err := ZipString(testData)
	if err != nil {
		t.Fatalf("failed to compress data: %v", err)
	}

	decompressedData, err := UnzipString(compressedData)
	if err != nil {
		t.Fatalf("failed to decompress data: %v", err)
	}

	if decompressedData != testData {
		t.Fatalf("decompressed data does not match original data: %v != %v", testData, decompressedData)
	}
}

func TestHuffmanFileData(t *testing.T) {
	RunFile("example.txt", t)
	RunFile("image.JPG", t)
}

func RunFile(targetPath string, t *testing.T) {
	_, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("failed to get target file info: %v", err)
	}

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed to read target file: %v", err)
	}

	testData := string(targetData)

	fmt.Printf("Compressing %s\n", targetPath)
	
	compressedData, err := ZipString(testData)
	if err != nil {
		t.Fatalf("failed to compress data: %v", err)
	}

	fileRatio := utils.NewFilesRatio(len(targetData), len(compressedData))

	fileRatio.PrintFileInfo()

	decompressedData, err := UnzipString(compressedData)
	if err != nil {
		t.Fatalf("failed to decompress data: %v", err)
	}

	if decompressedData != testData {
		t.Fatalf("decompressed data does not match original data: %v != %v", testData, decompressedData)
	}

	fileRatio.PrintCompressionRatio()
}

func ZipString(content string) (string, error) {
	files := []utils.File{
		{
			Name:    "file.txt",
			Content: []byte(content),
		},
	}
	compressedFile, err := Zip(files)
	if err != nil {
		return "", err
	}
	return string(compressedFile.Content), nil
}

func UnzipString(content string) (string, error) {
	file := utils.File{
		Name:    "compressed.sq",
		Content: []byte(content),
	}
	files, err := Unzip(file)
	if err != nil {
		return "", err
	}
	return string(files[0].Content), nil
}