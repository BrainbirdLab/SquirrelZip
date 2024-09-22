package hfc

import (
	"bytes"
	"file-compressor/utils"
	"path"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestCompressedData(t *testing.T) {
	path := []byte("compress_output/compressed.sq")
	//read
	file, err := os.Open(string(path))
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	output, err := os.Create("compress_output/output.txt")
	if err != nil {
		t.Fatalf("failed to create output file: %v", err)
	}

	freq := make(map[rune]int)
	getFrequencyMap(file, &freq)

	fmt.Printf("Frequency: %v\n", freq)

	codes, err := GetHuffmanCodes(&freq)
	if err != nil {
		t.Fatalf("failed to build huffman codes: %v", err)
	}

	fmt.Printf("Codes: %v\n", codes)

	_, err = compressData(file, output, codes)
	if err != nil {
		t.Fatalf("failed to compress data: %v", err)
	}
}

func TestFileRead(t *testing.T) {
	file, err := os.Open("input/example.txt")
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

func TestSmallString(t *testing.T) {
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

	err = CheckString(t, []byte("test_files/input/codes.txt"))
	if err != nil {
		t.Fatal(err)
	}

	err = CheckString(t, []byte("hi"))
	if err != nil {
		t.Fatal(err)
	}

	err = CheckString(t, []byte("Hello🥺😁 স্যার"))
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

	codes, err := GetHuffmanCodes(&freq)
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
	compLen, err := compressData(inputReader, compressedBuffer, codes)
	if err != nil {
		t.Fatalf("failed to compress: %v", err)
	}

	fmt.Printf("Compressed data length: %d\n", compLen)
	fmt.Printf("Compressed data: %s\n", compressedBuffer.Bytes())

	//decompress
	decompressedBytes := []byte{}
	decompressedBuffer := bytes.NewBuffer(decompressedBytes)

	fmt.Printf("CompLen: %d\n", compLen)
	
	err = decompressData(compressedBuffer, decompressedBuffer, codes, compLen)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	fmt.Printf("Decompressed data: %s\n", decompressedBuffer.Bytes())

	//compare original and decompressed data
	if bytes.Equal(testData, decompressedBuffer.Bytes()) == false {
		t.Fatal("original and decompressed data do not match")
	}

	return nil
}


func TestFile(t *testing.T) {
	RunFile("input/example.txt", t)
	//RunFile("image.JPG", t)
}

func RunFile(targetPath string, t *testing.T) {

	tempCompressPath := "compress_output"
	// Check if the output directory exists, create it if it doesn't
	if err := utils.MakeOutputDir(tempCompressPath); err != nil {
		t.Fatalf("failed to create output directory: %v", err)
	}

	targetStat, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("failed to get target file info: %v", err)
	}

	inputFile, err := os.Open(targetPath)
	if err != nil {
		t.Fatalf("failed to open target file: %v", err)
	}
	defer inputFile.Close()

	compressedFileName := path.Join(tempCompressPath, "compressed.sq")

	compressedFile, err := os.Create(compressedFileName)
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

	fmt.Printf("Compressed file: %s\n", compressedFileName)

	inputStat, err := inputFile.Stat()
	if err != nil {
		t.Fatalf("failed to get input file info: %v", err)
	}

	inputSize := inputStat.Size()
	
	compressedFile, err = os.Open(compressedFileName)
	if err != nil {
		t.Fatalf("failed to open compressed file again: %v", err)
	}

	stat, err := compressedFile.Stat()
	if err != nil {
		t.Fatalf("failed to get compressed file info: %v", err)
	}
	
	fileInfo := utils.NewFilesRatio(uint64(inputSize), uint64(stat.Size()))
	fileInfo.PrintFileInfo()
	
	// Decompress
	fileNames, err := Unzip(compressedFile, "decompress_output")
	if err != nil {
		t.Fatalf("failed to decompress file: %v", err)
	}
	
	if len(fileNames) != 1 {
		t.Fatalf("failed to decompress file: %v", err)
	}
	
	fmt.Printf("FileNames: %v\n", fileNames)
	
	fileInfo.PrintCompressionRatio()
	
	compressedFile.Close()

	//compare original and decompressed data
	compareFiles(targetPath, fileNames[0], t)
	//remove temp files
	//os.Remove(tempCompressPath)
}


func compareFiles(file1, file2 string, t *testing.T) {
	//compare original and decompressed files data
	originalFile, err := os.Open(file1)
	if err != nil {
		t.Fatalf("failed to open original file: %v", err)
	}
	defer originalFile.Close()

	decompressedFile, err := os.Open(file2)
	if err != nil {
		t.Fatalf("failed to open decompressed file: %v", err)
	}	
	defer decompressedFile.Close()

	originalReader := io.Reader(originalFile)
	decompressedReader := io.Reader(decompressedFile)

	originalBytes := make([]byte, 128)
	decompressedBytes := make([]byte, 128)

	for {
		n1, err1 := originalReader.Read(originalBytes)
		n2, err2 := decompressedReader.Read(decompressedBytes)

		if err1 != nil && err1 != io.EOF {
			t.Fatalf("failed to read original file: %v", err1)
		}

		if err2 != nil && err2 != io.EOF {
			t.Fatalf("failed to read decompressed file: %v", err2)
		}

		if n1 == 0 && n2 == 0 { // Reached EOF
			break
		}

		if bytes.Equal(originalBytes, decompressedBytes) == false {
			t.Fatal("original and decompressed data do not match")
		}
	}
}