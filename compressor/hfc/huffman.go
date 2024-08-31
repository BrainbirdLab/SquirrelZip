package hfc

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"errors"
	"file-compressor/utils"
	"fmt"
	"io"
	"os"
)

// Node represents a node in the Huffman tree.
type Node struct {
	char  rune  // Character stored in the node
	freq  int   // Frequency of the character
	left  *Node // Left child node
	right *Node // Right child node
}

// PriorityQueue implements heap.Interface and holds Nodes.
type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].freq == pq[j].freq {
		return pq[i].char < pq[j].char // Ensure deterministic order by comparing characters
	}
	return pq[i].freq < pq[j].freq
}

func (pq PriorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }

func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*Node))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// buildHuffmanTree builds the Huffman tree from character frequencies.
func buildHuffmanTree(freq *map[rune]int) (*Node, error) {
	if len(*freq) == 0 {
		return nil, errors.New("frequency map is empty")
	}

	pq := make(PriorityQueue, len(*freq))
	i := 0
	for char, f := range *freq {
		pq[i] = &Node{char: char, freq: f}
		i++
	}

	heap.Init(&pq)

	for len(pq) > 1 {
		left := heap.Pop(&pq).(*Node)
		right := heap.Pop(&pq).(*Node)
		internal := &Node{freq: left.freq + right.freq, left: left, right: right}
		heap.Push(&pq, internal)
	}

	return heap.Pop(&pq).(*Node), nil
}

func BuildHuffmanCodes(freq *map[rune]int) (map[rune]string, int8, error) {

	codes := make(map[rune]string)

	totalBits := int64(0)
	totalBytes := int8(0)

	node, err := buildHuffmanTree(freq)
	if err != nil {
		fmt.Printf("Error building Huffman tree: %v\n", err)
		return nil, 0, err
	}

	huffmanBuilder(node, "", &codes, freq, &totalBits)

	// Calculate the total number of bytes required to store the compressed data
	totalBytes = int8(totalBits / 8)

	return codes, totalBytes, nil
}

// huffmanBuilder builds Huffman codes for each character.
func huffmanBuilder(node *Node, prefix string, codes *map[rune]string, frequency *map[rune]int, totalBits *int64) {
	if node == nil {
		return
	}
	if node.left == nil && node.right == nil {
		(*codes)[node.char] = prefix
		// Calculate the number of bits used by this character
		*totalBits += int64(len(prefix) * (*frequency)[node.char])
		return
	}
	huffmanBuilder(node.left, prefix+"0", codes, frequency, totalBits)
	huffmanBuilder(node.right, prefix+"1", codes, frequency, totalBits)
}

func rebuildHuffmanTree(codes map[rune]string) *Node {

	root := &Node{}
	for char, code := range codes {
		node := root
		for _, bit := range code {
			if bit == '0' {
				if node.left == nil {
					node.left = &Node{}
				}
				node = node.left
			} else {
				if node.right == nil {
					node.right = &Node{}
				}
				node = node.right
			}
		}
		node.char = char
	}

	return root
}

func getFrequencyMap(input io.Reader, freq *map[rune]int) error {
	buf := make([]byte, 256)
	for {
		n, err := input.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading input: %w", err)
		}
		if n == 0 {
			break // EOF reached
		}

		for _, b := range buf[:n] {
			(*freq)[rune(b)]++
		}
	}

	return nil
}

func writeHuffmanCodes(codes map[rune]string, output io.Writer) error {
	// Write the number of codes
	if err := binary.Write(output, binary.LittleEndian, int32(len(codes))); err != nil {
		return fmt.Errorf("error writing number of codes: %w", err)
	}

	for char, code := range codes {
		if err := writeCodeEntry(char, code, output); err != nil {
			return fmt.Errorf("error writing code entry: %w", err)
		}
	}

	return nil
}

func writeCodeEntry(char rune, code string, output io.Writer) error {
	// Write the character
	if err := binary.Write(output, binary.LittleEndian, char); err != nil {
		return fmt.Errorf("error writing character: %w", err)
	}

	// Write the length of the code (in bits) to the file
	codeLen := int32(len(code))
	if err := binary.Write(output, binary.LittleEndian, codeLen); err != nil {
		return fmt.Errorf("error writing code length: %w", err)
	}

	// Convert the binary code string to a byte slice
	var buffer bytes.Buffer
	var value byte
	for i, bit := range code {
		value <<= 1 // Shift the value to the left by 1 bit
		if bit == '1' {
			value |= 1 // Set the least significant bit if the current bit is '1'
		}
		// if 8 bits have been read or it's the last bit, write the byte to the buffer
		if (i+1)%8 == 0 || i == len(code)-1 {
			buffer.WriteByte(value)
			value = 0
		}
	}

	// Write the code bytes to the file
	if _, err := output.Write(buffer.Bytes()); err != nil {
		return fmt.Errorf("error writing code: %w", err)
	}

	return nil
}

func readHuffmanCodes(input io.Reader, codes map[rune]string) error {
	numCodes, err := readNumOfCodes(input)
	if err != nil {
		return fmt.Errorf("error reading number of codes: %w", err)
	}

	for i := 0; i < int(numCodes); i++ {
		char, codeLen, err := readCode(input)
		if err != nil {
			return fmt.Errorf("error reading code: %w", err)
		}

		code, err := readCodeBytes(input, codeLen)
		if err != nil {
			return fmt.Errorf("error reading code bytes: %w", err)
		}

		codes[char] = code
	}

	return nil
}

func readNumOfCodes(input io.Reader) (int32, error) {
	var numCodes int32
	if err := binary.Read(input, binary.LittleEndian, &numCodes); err != nil {
		return 0, fmt.Errorf("error reading number of codes: %w", err)
	}

	return numCodes, nil
}

func readCode(input io.Reader) (rune, int32, error) {
	var char rune
	if err := binary.Read(input, binary.LittleEndian, &char); err != nil {
		return 0, 0, fmt.Errorf("error reading character: %w", err)
	}

	var codeLen int32
	if err := binary.Read(input, binary.LittleEndian, &codeLen); err != nil {
		return 0, 0, fmt.Errorf("error reading code length: %w", err)
	}

	return char, codeLen, nil
}

func readCodeBytes(input io.Reader, codeLen int32) (string, error) {
	var buffer bytes.Buffer
	bytesRead := int32(0)
	for bytesRead*8 < codeLen {
		var byteValue byte
		if err := binary.Read(input, binary.LittleEndian, &byteValue); err != nil {
			return "", fmt.Errorf("error reading code byte: %w", err)
		}

		for bit := 7; bit >= 0; bit-- {
			if bytesRead*8+int32(bit) < codeLen {
				if byteValue&(1<<bit) != 0 {
					buffer.WriteByte('1')
				} else {
					buffer.WriteByte('0')
				}
			}
		}
		bytesRead++
	}

	return buffer.String(), nil
}

func compressData(input io.Reader, output io.Writer, codes map[rune]string) error {
	var currentByte byte
	var bitCount uint8

	buf := make([]byte, 256)
	for {
		n, err := input.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading input: %w", err)
		}
		if n == 0 {
			break // EOF reached
		}

		if err := processByte(buf[:n], output, codes, &currentByte, &bitCount); err != nil {
			return fmt.Errorf("error processing byte: %w", err)
		}
	}
	if bitCount > 0 {
		// Pad the last byte with zeros
		fmt.Printf("Padding last byte with %d zeros\n", 8-bitCount)
		currentByte <<= 8 - bitCount
		if _, err := output.Write([]byte{currentByte}); err != nil {
			return fmt.Errorf("error writing compressed data: %w", err)
		}
	} else {
		fmt.Printf("No padding required\n")
		//write 8 
		if _, err := output.Write([]byte{byte(8)}); err != nil {
			return fmt.Errorf("error writing number of bits: %w", err)
		}
	}

	// Write the number of bits in the last byte
	if _, err := output.Write([]byte{byte(bitCount)}); err != nil {
		return fmt.Errorf("error writing number of bits: %w", err)
	}

	return nil
}

func processByte(buf []byte, output io.Writer, codes map[rune]string, currentByte *byte, bitCount *uint8) error {

	for _, b := range buf {
		char := rune(b)
		code, exists := codes[char]
		if !exists {
			return fmt.Errorf("no Huffman code for character %c", char)
		}

		for _, bit := range code {
			*currentByte <<= 1
			if bit == '1' {
				//left shift the current byte by 1 and set the least significant bit to 1
				*currentByte |= 1
			}
			*bitCount++
			if *bitCount == 8 {
				if _, err := output.Write([]byte{*currentByte}); err != nil {
					return fmt.Errorf("error writing compressed data: %w", err)
				}
				*currentByte = 0
				*bitCount = 0
			}
		}
	}

	return nil
}

func decompressData(input io.Reader, output io.Writer, codes map[rune]string) error {

	buf := make([]byte, 256)
	numOfBits := int8(0)
	lastByte := byte(0)
	remainingBits := byte(0)
	remainingBitsLen := int8(0)

	root := rebuildHuffmanTree(codes)

	for {
		buffLen, err := input.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		eof := make([]byte, 1)
		n, err := input.Read(eof)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			numOfBits = int8(buf[buffLen-1])
			//remove the last byte from the buffer
			buf = buf[:buffLen-1]
			buffLen-- // decrement the buffer length
			lastByte = buf[buffLen-1]
			buf = buf[:buffLen-1]
			buffLen--
		} else {
			fmt.Printf("adding eof byte to the buffer\n")
			//add the eof byte to the buffer
			buf = append(buf, eof...)
		}

		// compress the solid bytes (8 bits)
		if err := compressFullBits(buffLen, buf, output, root, &remainingBits, &remainingBitsLen); err != nil {
			return fmt.Errorf("error compressing full bits: %w", err)
		}

		// compress the remaining bits (incomplete bytes which are not multiples of 8)
		if n == 0 {
			return compressRemainingBits(remainingBits, remainingBitsLen, numOfBits, lastByte, output, root)
		}
	}
}

func compressRemainingBits(remainingBits byte, remainingBitsLen int8, numOfBits int8, lastByte byte, output io.Writer, root *Node) error {
	// include the last byte in the remaining bits
	for j := 7; j >= 8 - int(numOfBits); j-- {
		bit := (lastByte >> j) & 1
		remainingBits = (remainingBits << 1) | bit
		remainingBitsLen++
	}

	currentNode := root

	for i := remainingBitsLen - 1; i >= 0; i-- {
		
		bit := (remainingBits >> i) & 1

		if bit == 0 {
			currentNode = currentNode.left
		} else {
			currentNode = currentNode.right
		}

		if currentNode.left == nil && currentNode.right == nil {
			if _, err := output.Write([]byte{byte(currentNode.char)}); err != nil {
				return fmt.Errorf("error writing decompressed data: %w", err)
			}
			currentNode = root
		}
	}

	return nil
}

func compressFullBits(buffLen int, buf []byte, output io.Writer, root *Node, remainingBits *byte, remainingBitsLen *int8) error {
	currentNode := root
	// compress the solid bytes (8 bits)
	for i := 0; i < buffLen; i++ {
		for j := 7; j >= 0; j-- {
			bit := (buf[i] >> j) & 1
			*remainingBits = ((*remainingBits) << 1) | bit
			*remainingBitsLen++
			if bit == 0 {
				currentNode = currentNode.left
			} else {
				currentNode = currentNode.right
			}

			if currentNode.left == nil && currentNode.right == nil {
				if _, err := output.Write([]byte{byte(currentNode.char)}); err != nil {
					return fmt.Errorf("error writing decompressed data: %w", err)
				}
				currentNode = root
				// clear the remaining bits
				*remainingBits = 0
				*remainingBitsLen = 0
			}
		}
	}

	return nil
}

// Zip compresses data using Huffman coding and writes the compressed data to the output stream.
func Zip(files []utils.FileData, output io.Writer) error {

	freq := make(map[rune]int)
	codes := make(map[rune]string)

	//first, we need to get the frequency map
	for _, file := range files {

		reader := file.Reader
		//Get frequency map of the input file name
		for _, char := range file.Name {
			freq[char]++
		}

		//Get frequency map of the input data
		if err := getFrequencyMap(reader, &freq); err != nil {
			return fmt.Errorf("error generating frequency map: %w", err)
		}
		fmt.Printf("Frequency map length: %d\n", len(freq))

		//Build Huffman tree and codes
		root, err := buildHuffmanTree(&freq)
		if err != nil {
			return fmt.Errorf("error building Huffman tree: %w", err)
		}

		fmt.Println("Huffman tree built successfully")

		totalBits := int64(0)

		huffmanBuilder(root, "", &codes, &freq, &totalBits)

		fmt.Printf("codes length: %d\n", len(codes))

		// Write frequency map and Huffman codes to the output
		if err := writeHuffmanCodes(codes, output); err != nil {
			return fmt.Errorf("error writing Huffman codes: %w", err)
		}
	}

	fmt.Printf("Huffman codes written successfully\n")

	// Write the number of files
	if err := writeNumOfFiles(int32(len(files)), output); err != nil {
		return fmt.Errorf("error writing number of files: %w", err)
	}

	for _, file := range files {
		reader := file.Reader
		//Compress and write the file name
		if err := writeFileName(file.Name, output, codes); err != nil {
			return fmt.Errorf("error writing file name: %w", err)
		}
		//Compress and write the data
		if err := compressData(reader, output, codes); err != nil {
			return fmt.Errorf("error compressing data: %w", err)
		}
	}

	fmt.Printf("Data compressed successfully\n")

	return nil
}

func writeFileName(fileName string, output io.Writer, codes map[rune]string) error {
	// Write the file name length after compressing it
	fileNameBytes := []byte(fileName)
	nameBuf := bytes.NewBuffer(fileNameBytes)

	compressedNameBuf := bytes.NewBuffer([]byte{})

	fmt.Printf("Compressing file name: [%s], len(%d)\n", nameBuf.String(), nameBuf.Len())

	if err := compressData(nameBuf, compressedNameBuf, codes); err != nil {
		return fmt.Errorf("error compressing file name: %w", err)
	}

	fmt.Printf("Compressed name: [%s], len: (%d)\n", compressedNameBuf.String(), compressedNameBuf.Len())

	// write length of the file name buffer
	if err := binary.Write(output, binary.LittleEndian, uint64(compressedNameBuf.Len())); err != nil {
		return fmt.Errorf("error writing file name length: %w", err)
	}

	// write the compressed file name
	if _, err := output.Write(compressedNameBuf.Bytes()); err != nil {
		return fmt.Errorf("error writing file name: %w", err)
	}

	fmt.Printf("File name written successfully\n")

	return nil
}

func readNumOfFiles(input io.Reader) (int32, error) {
	var numOfFiles int32
	if err := binary.Read(input, binary.LittleEndian, &numOfFiles); err != nil {
		return 0, fmt.Errorf("error reading number of files: %w", err)
	}

	fmt.Printf("Number of files read: %d\n", numOfFiles)

	return numOfFiles, nil
}

func writeNumOfFiles(numOfFiles int32, output io.Writer) error {

	if err := binary.Write(output, binary.LittleEndian, numOfFiles); err != nil {
		return fmt.Errorf("error writing number of files: %w", err)
	}

	fmt.Printf("Number of files written: %d\n", numOfFiles)

	return nil
}

func readFileName(input io.Reader, codes map[rune]string) (string, error) {
	var nameLen uint64
	if err := binary.Read(input, binary.LittleEndian, &nameLen); err != nil {
		return "", fmt.Errorf("error reading file name length: %w", err)
	}

	fmt.Printf("File name length read: %d\n", nameLen)

	buf := make([]byte, nameLen)
	if _, err := input.Read(buf); err != nil {
		return "", fmt.Errorf("error reading file name: %w", err)
	}

	compressedFilename := bytes.NewBuffer(buf)

	fmt.Printf("File name buffer read: %s\n", compressedFilename.String())

	nameBuffer := bytes.NewBuffer([]byte{})
	if err := decompressData(compressedFilename, nameBuffer, codes); err != nil {
		return "", fmt.Errorf("error decompressing file name: %w", err)
	}

	name := nameBuffer.String()

	fmt.Printf("File name read: %s\n", name)

	return name, nil
}

// Unzip decompresses data using Huffman coding and writes the decompressed data to the output stream.
func Unzip(input io.Reader, outputPath string) ([]string, error) {

	codes := make(map[rune]string)
	if err := readHuffmanCodes(input, codes); err != nil {
		return nil, fmt.Errorf("error reading Huffman codes: %w", err)
	}

	numOfFiles, err := readNumOfFiles(input)
	if err != nil {
		return nil, fmt.Errorf("error reading number of files: %w", err)
	}

	filePaths := []string{}

	for i := 0; i < int(numOfFiles); i++ {

		// get the file name
		fileName, err := readFileName(input, codes)
		if err != nil {
			return nil, fmt.Errorf("error reading file name: %w", err)
		}

		fileName = outputPath + fileName

		// if output dir doesn't exist, create it
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			err := os.MkdirAll(outputPath, 0755)
			if err != nil {
				return nil, fmt.Errorf("error creating output directory: %w", err)
			}
		}

		// writer
		outputFile, err := os.Create(fileName)
		if err != nil {
			return nil, fmt.Errorf("error creating output file: %w", err)
		}

		err = decompressData(input, outputFile, codes)
		if err != nil {
			return nil, fmt.Errorf("error decompressing data: %w", err)
		}

		outputFile.Close()

		fmt.Printf("File decompressed successfully: %s\n", fileName)

		filePaths = append(filePaths, fileName)
	}

	return filePaths, nil
}
