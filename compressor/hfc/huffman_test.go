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

func TestSmallStringHuffman(t *testing.T) {
	err := CheckString(t, []byte("Hello"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestBitEncoding(t *testing.T) {

	err := CheckString(t, []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."))
	if err != nil {
		t.Fatal(err)
	}

	err = CheckString(t, []byte("examination"))
	if err != nil {
		t.Fatal(err)
	}

	err = CheckString(t, []byte("hi"))
	if err != nil {
		t.Fatal(err)
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

	codes, err := BuildHuffmanCodes(&freq)
	if err != nil {
		t.Fatalf("failed to build huffman codes: %v", err)
	}

	//reset the seek
	//inputReader.Seek(0, io.SeekStart)
	if len(codes) == 0 {
		return fmt.Errorf("failed to build huffman codes")
	}

	compressedBytes := []byte{}
	compressedBuffer := bytes.NewBuffer(compressedBytes)

	//compress
	err = compressData(inputReader, compressedBuffer, codes)
	if err != nil {
		t.Fatalf("failed to compress: %v", err)
	}

	//decompress
	decompressedBytes := []byte{}
	decompressedBuffer := bytes.NewBuffer(decompressedBytes)
	
	err = decompressData(compressedBuffer, decompressedBuffer, codes)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	//compare original and decompressed data
	if bytes.Equal(testData, decompressedBuffer.Bytes()) == false {
		t.Fatal("original and decompressed data do not match")
	}

	return nil
}


func TestHuffmanFileData(t *testing.T) {
	RunFile("example.txt", t)
	//RunFile("image.JPG", t)
}

func RunFile(targetPath string, t *testing.T) {

	tempCompressPath := "compressed.sq"

	targetStat, err := os.Stat(targetPath)
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

	inputFileData := utils.FileData{
		Name: targetPath,
		Size: targetStat.Size(),
		Reader: inputFile,
	}
	
	// Compress
	err = Zip([]utils.FileData{inputFileData}, compressedFile)
	if err != nil {
		t.Fatalf("failed to compress file: %v", err)
	}

	compressedFile.Close()

	

	inputStat, err := inputFile.Stat()
	if err != nil {
		t.Fatalf("failed to get input file info: %v", err)
	}

	inputSize := inputStat.Size()
	
	compressedFile, _ = os.Open(tempCompressPath)
	
	compressedStat, err := compressedFile.Stat()
	if err != nil {
		t.Fatalf("failed to get compressed file info: %v", err)
	}

	
	compressedSize := compressedStat.Size()
	
	fileInfo := utils.NewFilesRatio(inputSize, compressedSize)
	fileInfo.PrintFileInfo()
	
	// Decompress
	fileNames, err := Unzip(compressedFile, "output")
	if err != nil {
		t.Fatalf("failed to decompress file: %v", err)
	}
	
	if len(fileNames) != 1 {
		t.Fatalf("failed to decompress file: %v", err)
	}
	
	fmt.Printf("FileNames: %v\n", fileNames)
	
	fileInfo.PrintCompressionRatio()
	
	compressedFile.Close()
	
	//remove temp files
	//os.Remove(tempCompressPath)
}
