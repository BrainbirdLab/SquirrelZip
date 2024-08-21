package huffmanCoding

import (
	"file-compressor/utils"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestHuffman(t *testing.T) {
	testData := "Hello, World! Hello are you doing today? I am doing great! How about you? This is a test message to test the huffman coding algorithm. I hope it works well."

	// Create input file and write test data
	inputFile, err := os.Create("test.txt")
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer inputFile.Close()

	_, err = inputFile.WriteString(testData)
	if err != nil {
		t.Fatalf("failed to write test data to file: %v", err)
	}

	// Close inputFile before opening compressedFile
	inputFile.Close()

	// Create compressed file
	compressedFile, err := os.Create("test.txt.huffman")
	if err != nil {
		t.Fatalf("failed to create compressed file: %v", err)
	}
	defer compressedFile.Close()

	// Open input file again for reading
	inputFile, err = os.Open("test.txt")
	if err != nil {
		t.Fatalf("failed to open input file: %v", err)
	}
	defer inputFile.Close()

	// Compress
	err = Zip(inputFile, compressedFile)
	if err != nil {
		t.Fatalf("failed to compress file: %v", err)
	}

	// Read compressed data
	compressedFile, err = os.Open("test.txt.huffman")
	if err != nil {
		t.Fatalf("failed to open compressed file: %v", err)
	}
	defer compressedFile.Close()

	cData, err := io.ReadAll(compressedFile)
	if err != nil {
		t.Fatalf("failed to read compressed data: %v", err)
	}

	fileInfo := utils.NewFilesRatio(len(testData), len(cData))
	fileInfo.PrintFileInfo()

	// Create decompressed file
	decompressedFile, err := os.Create("test.txt.decompressed")
	if err != nil {
		t.Fatalf("failed to create decompressed file: %v", err)
	}
	defer decompressedFile.Close()

	// Open compressed file again for decompression
	compressedFile, err = os.Open("test.txt.huffman")
	if err != nil {
		t.Fatalf("failed to open compressed file: %v", err)
	}
	defer compressedFile.Close()

	// Decompress
	err = Unzip(compressedFile, decompressedFile)
	if err != nil {
		t.Fatalf("failed to decompress file: %v", err)
	}

	// Reset the decompressed file pointer to the beginning for reading
	decompressedFile, err = os.Open("test.txt.decompressed")
	if err != nil {
		t.Fatalf("failed to open decompressed file: %v", err)
	}
	defer decompressedFile.Close()

	// Read decompressed data
	dData, err := io.ReadAll(decompressedFile)
	if err != nil {
		t.Fatalf("failed to read decompressed data: %v", err)
	}

	fmt.Printf("Original data:%s\n", testData)
	fmt.Printf("Decompressed data:%s\n", dData)


	if string(dData) != testData {
		t.Fatalf("decompressed data does not match original data\nOrinal length: %d, Decompressed length: %d", len(testData), len(dData))
	}

	fileInfo.PrintCompressionRatio()
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

	inputFile, err := os.Open(targetPath)
	if err != nil {
		t.Fatalf("failed to open target file: %v", err)
	}
	defer inputFile.Close()

	compressedFile, err := os.Create(targetPath + ".huffman")
	if err != nil {
		t.Fatalf("failed to create compressed file: %v", err)
	}
	defer compressedFile.Close()

	// Compress
	err = Zip(inputFile, compressedFile)
	if err != nil {
		t.Fatalf("failed to compress file: %v", err)
	}

	decompressedFile, err := os.Create(targetPath + ".decompressed")
	if err != nil {
		t.Fatalf("failed to create decompressed file: %v", err)
	}
	defer decompressedFile.Close()

	// Decompress
	err = Unzip(compressedFile, decompressedFile)
	if err != nil {
		t.Fatalf("failed to decompress file: %v", err)
	}
}