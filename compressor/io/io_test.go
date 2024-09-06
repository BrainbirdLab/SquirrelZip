package io

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestBitsCombination(t *testing.T) {
	byte1 := byte(0b10100000)
	byte2 := byte(0b00010000)
	expected := []byte{0b10100001}
	// take first 4 bits from byte1 and last 4 bits from byte2
	result := concatenateBits(byte1, 4, byte2, 4)
	if !bytes.Equal(result, expected) {
		t.Fatalf("Test case 1 failed: expected %v, got %v", expected, result)
	}

	byte1 = byte(0b10100000)
	byte2 = byte(0b00010000)
	expected = []byte{0b10100000, 0b10000000}
	// take first 5 bits from byte1 and last 5 bits from byte2
	result = concatenateBits(byte1, 5, byte2, 5)
	if !bytes.Equal(result, expected) {
		t.Fatalf("Test case 2 failed: expected %v, got %v", expected, result)
	}

	byte1 = byte(0b00000000)
	byte2 = byte(0b00100000)
	expected = []byte{0b00000100}
	// take first 6 bits from byte1 and last 2 bits from byte2
	result = concatenateBits(byte1, 3, byte2, 3)
	if !bytes.Equal(result, expected) {
		t.Fatalf("Test case 3 failed: expected %v, got %v", expected, result)
	}
}

func concatenateBits(byte1 byte, bits1 int, byte2 byte, bits2 int) []byte {
    // Create a byte slice to store the result
	bitsNeeded := bits1 + bits2
	result := make([]byte, int(math.Ceil(float64(bitsNeeded) / 8)))
	currentByte := byte(0)
	bitIndex := 0
	arrayIndex := 0
	// Iterate over the bits in the first byte
	for i := 0; i < bits1; i++ {
		// Shift the current byte to the left
		currentByte <<= 1
		// Get the value of the current bit
		bit := (byte1 >> (7 - i)) & 1
		// Set the least significant bit of the current byte
		currentByte |= bit
		// Increment the bit index
		bitIndex++

		// If the current byte is full, add it to the result
		if bitIndex == 8 {
			result[arrayIndex] = currentByte
			arrayIndex++
			bitIndex = 0
			currentByte = 0
		}
	}

	// Iterate over the bits in the second byte
	for i := 0; i < bits2; i++ {
		// Shift the current byte to the left
		currentByte <<= 1
		// Get the value of the current bit
		bit := (byte2 >> (7 - i)) & 1
		// Set the least significant bit of the current byte
		currentByte |= bit
		// Increment the bit index
		bitIndex++

		// If the current byte is full, add it to the result
		if bitIndex == 8 {
			result[arrayIndex] = currentByte
			arrayIndex++
			bitIndex = 0
			currentByte = 0
		}
	}

	// If there are any bits left in the current byte, add it to the result
	if bitIndex > 0 {
		//pad the remaining bits with 0 if needed
		currentByte <<= uint(8 - bitIndex)
		result[arrayIndex] = currentByte
	}

    return result
}

func TestChunkReader(t *testing.T) {
	//read data without advancing. Reader type is io.Reader (os.File or bytes.Buffer)
	buffer := []byte("hi mom..!")
	reader := bytes.NewReader(buffer)
	
	// [N - 2 byte buffer][last byte][last byte's bit count]...[N - 2 byte buffer][last byte][last byte's bit count]
	outputArray := []byte{}
	outputWriter := bytes.NewBuffer(outputArray)

	err := ReadData(reader, outputWriter)
	if err != nil {
		t.Fatalf("failed to read data: %v", err)
	}

	fmt.Printf("Output: [%s]\n", outputWriter.String())
	fmt.Printf("Output: [%v]\n", outputWriter.String())
	fmt.Printf("Output: [%08b]\n", outputWriter.Bytes())

	//check if the output matches the input
	if outputWriter.String() != string(buffer) {
		t.Fatalf("Data mismatch expected length: %d, output length %d\n", len(buffer), len(outputWriter.String()))
	}
}

func ReadData(reader io.Reader, writer io.Writer) error {

	N := 4 //chunk size

	lastByte := make([]byte, 1)
	lastByteCount := make([]byte, 1)

	loopFlag := 0

	for {

		readBuffer := make([]byte, N)
		n, err := reader.Read(readBuffer)
		if err != nil && err != io.EOF {
			return err
		}

		fmt.Printf("Read [%s] in current iteration.\n", readBuffer)

		if n == 0 {
			// add remaining last byte and bit count to the output
			/*
			output = append(output, lastByte...)
			output = append(output, lastByteCount...)
			*/
			fmt.Printf("Last byte: [%s] of length: %d\n", lastByte, len(lastByte))
			fmt.Printf("Last byte count: [%s] of length: %d\n", lastByteCount, len(lastByteCount))
			finalData := append(lastByte, lastByteCount...)
			writer.Write(finalData)
			fmt.Printf("Final data: [%s] of length: %d written.\n", finalData, len(finalData))
			break
		} else {

			if loopFlag != 0 {
				n += 2
				//add the last byte and bit count to the readBuffer
				leftOverData := append(lastByte, lastByteCount...)
				readBuffer = append(leftOverData, readBuffer...)
				fmt.Printf("Left over data: [%s] of length: %d\n", leftOverData, len(leftOverData))
			}

			// add the last byte from the chunk to the lastByte buffer
			lastByte = readBuffer[n-2:n-1]
			// add the last byte's bit count to the lastByteCount buffer
			lastByteCount = readBuffer[n-1:n]
			// remove last 2 bytes from the chunk
			readBuffer = readBuffer[:n-2]

			// add the chunk
			//output = append(output, chunk...)
			writer.Write(readBuffer)
			fmt.Printf("Chunk: [%s] of length: %d written.\n", readBuffer, len(readBuffer))
		}

		loopFlag = 1
	}

	return nil
}

func TestArray(t *testing.T) {
	arr := make([]byte, 2)
	arr[0] = 1
	arr[1] = 2
	fmt.Println(arr)
	//add more elements to the array
	arr = append(arr, 3)
	fmt.Println(arr)
}

func WriteMap(data map[string]int, key string, value int) {
	data[key] = value
}

func TestWriteMap(t *testing.T) {
	data := make(map[string]int)
	WriteMap(data, "key", 10)
	fmt.Println(data)
	WriteMap(data, "key", 20)
	fmt.Println(data)
}

func TestRecursiveFolderRead(t *testing.T) {
	//get all filenames from a dir. include subdirectories
	target := "test_files"

	files := make([]string, 0)

	//use walk to get all files in the directory
	err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk directory: %v", err)
	}

	fmt.Printf("Files found: %v\n", files)
}

func TestRecursiveFolderWrite(t *testing.T) {
	path := "recursive_test/test1/test2/test3/file.txt"
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	file.Close()
	//delete the recursive_test directory
	defer os.RemoveAll("recursive_test")
}