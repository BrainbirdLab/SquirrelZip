package hfc

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

func TestBinaryWrite(t *testing.T) {
	// Open a file for writing
	file, err := os.Create("temp/binary.txt")
	if err != nil {
		t.Fatalf("err creating file: %v", err)
	}
	defer file.Close()

	// Binary data of 47 (binary "101111")
	binaryData := "00101111"

	// Convert the binary string to an integer
	var value byte
	for i := 0; i < len(binaryData); i++ {
		value <<= 1 // Shift the value to the left by 1 bit
		if binaryData[i] == '1' {
			value |= 1 // Set the least significant bit if the current bit is '1'
		}
	}

	// Write the byte (which is 47 in this case) to the file
	_, err = file.Write([]byte{value})
	if err != nil {
		t.Fatalf("err writing to file: %v", err)
	}
}

func TestIO(t *testing.T) {
	// Example data source
	data := []byte("This is a stream of data that we're going to read in chunks.")
	reader := bytes.NewReader(data)

	// Buffer to hold each chunk of data
	buffer := make([]byte, 16) // Read 16 bytes at a time

	for {
		// Read data into the buffer
		n, err := reader.Read(buffer)
		
		// Check if the end of the stream is reached
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// Process the chunk of data read
		fmt.Printf("Read %d bytes: %s\n", n, buffer[:n])
	}

	fmt.Println("Finished reading the stream.")
}

func TestSeek(t *testing.T) {

	filePath := "temp/str.txt"
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("err reading file: %v", err)
	}
	defer file.Close()

	// Seek to the beginning of the file
	WriteAt(file, -19, "HEHE")
}

func WriteAt(writer io.Writer, n int64, data string) {
	// Check if the data source implements the Seeker interface
	seeker, ok := writer.(io.Seeker)
	if !ok {
		fmt.Println("Data source does not support seeking.")
		return
	}

	// Seek to the specified position
	_, err := seeker.Seek(n, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking:", err)
		return
	}

	fmt.Println("Seek successful.")

	// Write data at the current position
	_, err = writer.Write([]byte(data))
	if err != nil {
		fmt.Println("Error writing:", err)
		return
	}

	fmt.Println("Write successful.")

	// Seek back to the end of the data source
	_, err = seeker.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Println("Error seeking:", err)
		return
	}
}
