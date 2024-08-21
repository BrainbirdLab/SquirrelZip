package hfc

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"testing"
)

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
