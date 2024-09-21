package main

import (
	"os"
	"testing"
)

func TestIO(t *testing.T) {
	Init(t, "huffman")
	Init(t, "arithmetic")
}

func Init(t *testing.T, algo string) {

	t.Log("TestIO: Testing IO operations with algorithm: ", algo)
	zipOutputDir := "test_zip"
	unzipOutputDir := "test_unzip"

	clean(zipOutputDir, unzipOutputDir)

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