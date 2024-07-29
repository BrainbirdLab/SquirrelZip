package main

import (
	"file-compressor/compressor"
	"file-compressor/utils"
	"os"
	"testing"
)

func TestIO(t *testing.T) {
	t.Log("TestIO")
	inputDir := "test_files"
	zipOutputDir := "test_zip"
	unzipOutputDir := "test_unzip"

	testPassword := ""

	inputFilesNames, err := utils.GetAllFileNamesFromDir(&inputDir)
	if err != nil {
		t.Fatalf("failed to read test files from %s: %v", inputDir, err)
	}
	//compress test files
	err = compressor.Compress(inputFilesNames, zipOutputDir, testPassword)
	if err != nil {
		t.Fatalf("failed to compress test files: %v", err)
	}

	//decompress test files
	err = compressor.Decompress(zipOutputDir + "/" + "compressed.sq", unzipOutputDir, testPassword)
	if err != nil {
		t.Fatalf("failed to decompress test files: %v", err)
	}

	//clean up
	os.RemoveAll(zipOutputDir)
	os.RemoveAll(unzipOutputDir)
}