package main

import (
	"os"
	"testing"

	"file-compressor/compressor"
	"file-compressor/utils"
)

func TestIO(t *testing.T) {
	Init(t, "huffman")
	Init(t, "arithmetic")
}

func Init(t *testing.T, algo string) {

	t.Log("TestIO: Testing IO operations with algorithm: ", algo)
	inputDir := "test_files"
	zipOutputDir := "test_zip"
	unzipOutputDir := "test_unzip"

	clean(zipOutputDir, unzipOutputDir)

	testPassword := ""

	inputFilesNames, err := utils.GetAllFileNamesFromDir(&inputDir)
	if err != nil {
		t.Fatalf("failed to read test files from %s: %v", inputDir, err)
	}
	//compress test files
	err = compressor.Compress(inputFilesNames, zipOutputDir, testPassword, algo)
	if err != nil {
		t.Fatalf("failed to compress test files: %v", err)
	}

	//decompress test files
	err = compressor.Decompress(zipOutputDir + "/" + "compressed.sq", unzipOutputDir, testPassword)
	if err != nil {
		t.Fatalf("failed to decompress test files: %v", err)
	}

	defer func ()  {		
		//clean up
		t.Log("Cleanup")
		clean(zipOutputDir, unzipOutputDir)
	}()
}

func clean(zipOutputDir, unzipOutputDir string) {
	os.RemoveAll(zipOutputDir)
	os.RemoveAll(unzipOutputDir)
}