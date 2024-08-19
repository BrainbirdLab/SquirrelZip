package compressor

import (
	"os"
	"testing"

	"file-compressor/compressor/arithmetic"
	hc "file-compressor/compressor/huffmanCoding"
	"file-compressor/utils"
)

func TestCompress(t *testing.T) {
	Init("huffman", t)
	Init("arithmetic", t)
}

func Init(algo string, t *testing.T) {

	testFilesDir := "./../test_files"
	testFiles := []utils.File{}

	currDir, _ := os.Getwd()

	//read test files
	allFileNames, err := utils.GetAllFileNamesFromDir(&testFilesDir)

	if err != nil {
		t.Fatalf("failed to read test files from %s: %v", currDir, err)
	}

	originalSize := int64(0)

	ReadAllFilesConcurrently(allFileNames, &testFiles, &originalSize)

	err = CheckCompressionAlgorithm(algo)
	if err != nil {
		t.Fatal(err)
	}

	compressedFile := utils.File{}

	switch utils.Algorithm(algo) {
	case utils.HUFFMAN:
		compressedFile, err = hc.Zip(testFiles)
	case utils.ARITHMETIC:
		compressedFile, err = arithmetic.Zip(testFiles)
	}

	if err != nil {
		t.Fatalf("failed to compress test files: %v", err)
	}

	WriteCompressionAlgorithm(algo, &compressedFile)

	usedAlgorithm, err := ReadCompressionAlgorithm(&compressedFile)
	if err != nil {
		t.Fatalf("failed to read compression algorithm: %v", err)
	}

	t.Logf("Used compression algorithm: %v\n", usedAlgorithm)

	decompressedFiles := []utils.File{}
	
	switch usedAlgorithm {
	case utils.HUFFMAN:
		decompressedFiles, err = hc.Unzip(compressedFile)
	case utils.ARITHMETIC:
		decompressedFiles, err = arithmetic.Unzip(compressedFile)
	}

	if err != nil {
		t.Fatalf("failed to decompress test files [%s]: %v", usedAlgorithm, err)
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

	cmfPath := testFilesDir + "/" + "test.cmf"

	//store the compressed file
	err = os.WriteFile(cmfPath, compressedFile.Content, 0644)
	if err != nil {
		t.Fatalf("failed to save compressed file: %v", err)
	}

	//get the file size
	compressedFileInfo, err := os.Stat(cmfPath)
	if err != nil {
		t.Fatalf("failed to get compressed file info: %v", err)
	}

	compressedSize := compressedFileInfo.Size()

	// Calculate compression ratio
	compressionRatio := float64(compressedSize) / float64(originalSize)

	t.Logf("Compression ratio [%s]: %f%%", algo, compressionRatio)

	//delete the output file
	os.RemoveAll(cmfPath)
	t.Logf("Deleted compressed file: %s", cmfPath)
}