package arithmetic

import (
	"file-compressor/utils"
	"os"
	"testing"
)

func TestArithmetic(t *testing.T) {
	
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

	defer func() {
		//cleanup
		os.RemoveAll("./test_zip")
	}()

	return string(files[0].Content), nil
}