package hfc

import (
	"bytes"
	"file-compressor/utils"
	//"encoding/binary"
	//"file-compressor/utils"
	"fmt"
	"io"
	"os"

	//"fmt"
	//"io"
	//"os"
	"testing"
)

func TestFileRead(t *testing.T) {
	file, err := os.Open("temp/test.txt")
	if err != nil {
		t.Fatalf("err reading file: %v", err)
	}

	fbuf := io.Reader(file)

	fbytes := make([]byte, 128)

	for {
		n, err := fbuf.Read(fbytes)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading:", err)
			break
		}

		if n == 0 { // Reached EOF
			break
		}
	}
}

func TestHuffman(t *testing.T) {
	Valid := []byte("Hello, World! Hello are you doing today? I am doing great! How about you? This is a test message to test the huffman coding algorithm. I hope it works well.")
	
	err := CheckString(t, Valid)
	if err != nil {
		t.Fatalf("failed to compress/decompress: %v", err)
	}

	empty := []byte("")

	err = CheckString(t, empty)
	if err == nil {
		t.Fatalf("failed to compress/decompress: %v", err)
	}
}

func CheckString(t *testing.T, testData []byte) error {

	if len(testData) == 0 {
		//expected error
		return fmt.Errorf("empty data")
	}

	inputReader := bytes.NewReader(testData)

	//test huffman
	freq := make(map[rune]int)
	getFrequencyMap(inputReader, &freq)
	//reset the seek
	inputReader.Seek(0, io.SeekStart)

	if len(freq) == 0 {
		return fmt.Errorf("failed to get frequency map")
	}

	root, err := buildHuffmanTree(&freq)
	if err != nil {
		t.Fatalf("failed to build huffman tree: %v", err)
	}

	codes := make(map[rune]string)
	buildHuffmanCodes(root, "", codes)

	//reset the seek
	//inputReader.Seek(0, io.SeekStart)
	if len(codes) == 0 {
		return fmt.Errorf("failed to build huffman codes")
	}

	compressedBytes := []byte{}
	compressedBuffer := bytes.NewBuffer(compressedBytes)
	
	//compress
	err = compressData(inputReader, compressedBuffer, &codes)
	if err != nil {
		t.Fatalf("failed to compress: %v", err)
	}

	compressedBufferSize := len(compressedBuffer.Bytes())

	//decompress
	decompressedBytes := []byte{}
	decompressedBuffer := bytes.NewBuffer(decompressedBytes)
	
	err = decompressData(compressedBuffer, decompressedBuffer, root)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}


	//compare original and decompressed data
	if decompressedBuffer.String() != string(testData) {
		t.Fatalf("original and decompressed data do not match")
	} else {
		fmt.Printf("Original length: %d, Compressed length: %d, Decompressed length: %d\n", len(testData), compressedBufferSize, len(decompressedBuffer.Bytes()))
	}

	return nil
}


func TestHuffmanFileData(t *testing.T) {
	RunFile("example.txt", t)
	//RunFile("image.JPG", t)
}

func RunFile(targetPath string, t *testing.T) {

	tempCompressPath := targetPath + ".huffman.txt"
	tempDecompressPath := targetPath + ".decompressed.txt"

	_, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("failed to get target file info: %v", err)
	}

	inputFile, err := os.Open(targetPath)
	if err != nil {
		t.Fatalf("failed to open target file: %v", err)
	}
	defer inputFile.Close()

	compressedFile, err := os.Create(tempCompressPath)
	if err != nil {
		t.Fatalf("failed to create compressed file: %v", err)
	}
	
	// Compress
	err = Zip(inputFile, compressedFile)
	if err != nil {
		t.Fatalf("failed to compress file: %v", err)
	}

	inputStat, err := inputFile.Stat()
	if err != nil {
		t.Fatalf("failed to get input file info: %v", err)
	}

	inputSize := inputStat.Size()
	
	
	compressedStat, err := compressedFile.Stat()
	if err != nil {
		t.Fatalf("failed to get compressed file info: %v", err)
	}
	compressedFile.Close()

	compressedSize := compressedStat.Size()

	fileInfo := utils.NewFilesRatio(inputSize, compressedSize)
	fileInfo.PrintFileInfo()

	decompressedFile, err := os.Create(tempDecompressPath)
	if err != nil {
		t.Fatalf("failed to create decompressed file: %v", err)
	}

	
	compressedFile, err = os.Open(tempCompressPath)
	if err != nil {
		t.Fatalf("failed to open compressed file: %v", err)
	}

	// Decompress
	err = Unzip(compressedFile, decompressedFile)
	if err != nil {
		t.Fatalf("failed to decompress file: %v", err)
	}
	
	compressedFile.Close()
	decompressedFile.Close()

	fileInfo.PrintCompressionRatio()

	//remove temp files
	os.Remove(tempCompressPath)
	os.Remove(tempDecompressPath)
}
