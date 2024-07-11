package compressor

import (
	"os"
	"testing"

	"file-compressor/utils"
)

func TestCompress(t *testing.T) {
	testFilesDir := "./../test_files"
	testFiles := []utils.File{}

	currDir, _ := os.Getwd()

	t.Logf("Current working directory: %v\n", currDir)

	//read test files
	allFileNames, err := utils.GetAllFileNamesFromDir(&testFilesDir)

	if err != nil {
		t.Fatalf("failed to read test files from %s: %v", currDir, err)
	}

	originalSize := int64(0)

	ReadAllFilesConcurrently(allFileNames, &testFiles, &originalSize)

	//compress test files
	compressedFile, err := Zip(testFiles)

	if err != nil {
		t.Fatalf("failed to compress test files: %v", err)
	}

	//decompress test files
	decompressedFiles, err := Unzip(compressedFile)

	if err != nil {
		t.Fatalf("failed to decompress test files: %v", err)
	}

	//compare original and decompressed files
	for i, file := range testFiles {
		if file.Name != decompressedFiles[i].Name {
			t.Fatalf("file name mismatch: %v != %v", file.Name, decompressedFiles[i].Name)
		}

		if string(file.Content) != string(decompressedFiles[i].Content) {
			t.Fatalf("file content mismatch: %v != %v", string(file.Content), string(decompressedFiles[i].Content))
		}
	}
}
