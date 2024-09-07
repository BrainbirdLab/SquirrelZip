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
	files := []utils.FileData{
		{
			Name:    "file.txt",
		},
	}
	_, err := Zip(files)
	if err != nil {
		return "", err
	}
	return "", nil
}

func UnzipString(content string) (string, error) {
	file := utils.FileData{
		Name:    "compressed.sq",
	}
	
	_, err := Unzip(file)
	if err != nil {
		return "", err
	}

	defer func() {
		//cleanup
		os.RemoveAll("./test_zip")
	}()

	return "", nil
}