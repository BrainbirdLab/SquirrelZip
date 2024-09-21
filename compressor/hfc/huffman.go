package hfc

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	ec "file-compressor/errorConstants"
	"file-compressor/utils"
)

const (
	BUFFER_SIZE = 256
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

func GetHuffmanCodes(freq *map[rune]int) (map[rune]string, error) {

	codes := make(map[rune]string)

	node, err := buildHuffmanTree(freq)
	if err != nil {
		return nil, err
	}

	huffmanBuilder(node, "", &codes, freq)

	return codes, nil
}

// huffmanBuilder builds Huffman codes for each character.
func huffmanBuilder(node *Node, prefix string, codes *map[rune]string, frequency *map[rune]int) {
	if node == nil {
		return
	}
	if node.left == nil && node.right == nil {
		(*codes)[node.char] = prefix
		return
	}
	huffmanBuilder(node.left, prefix+"0", codes, frequency)
	huffmanBuilder(node.right, prefix+"1", codes, frequency)
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
			return fmt.Errorf(ec.BUFFER_READ_ERROR, err)
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

func WriteHuffmanCodes(file io.Writer, codes map[rune]string) error {
	//write the number of codes
	binary.Write(file, binary.LittleEndian, uint16(len(codes)))

	for k, v := range codes {
		//write the rune
		binary.Write(file, binary.LittleEndian, uint32(k))
		//write the length of the code
		binary.Write(file, binary.LittleEndian, byte(len(v))) // bit length of the code
		//write the code
		completeByte := byte(0)
		completeByteIndex := 0
		for i := 0; i < len(v); i++ {
			if v[i] == '1' {
				completeByte |= 1 << uint8(7-completeByteIndex)
			}
			completeByteIndex++
			if completeByteIndex == 8 {
				binary.Write(file, binary.LittleEndian, completeByte)
				completeByte = 0
				completeByteIndex = 0
			}
		}

		// Write the last byte if it is not complete
		if completeByteIndex > 0 {
			binary.Write(file, binary.LittleEndian, completeByte)
		}
	}

	return nil
}

func ReadHuffmanCodes(file io.Reader) (map[rune]string, error) {
	codes := make(map[rune]string)

	// Read the number of codes (4 bytes, uint32)
	var numCodes uint16
	if err := binary.Read(file, binary.LittleEndian, &numCodes); err != nil {
		return nil, fmt.Errorf(ec.FILE_READ_ERROR, err)
	}

	// Read each code
	for i := uint16(0); i < numCodes; i++ {
		// Read the rune
		var r uint32
		if err := binary.Read(file, binary.LittleEndian, &r); err != nil {
			return nil, fmt.Errorf(ec.FILE_READ_ERROR, err)
		}

		// Read the length of the Huffman code (1 byte)
		var codeLen byte
		if err := binary.Read(file, binary.LittleEndian, &codeLen); err != nil {
			return nil, fmt.Errorf(ec.FILE_READ_ERROR, err)
		}

		// Read the code bits (packed into bytes)
		codeBits := make([]byte, (codeLen+7)/8) // (codeLen+7)/8 ensures enough space to hold all bits
		if _, err := file.Read(codeBits); err != nil {
			return nil, fmt.Errorf(ec.FILE_READ_ERROR, err)
		}

		// Convert the code bits back into a string of '0's and '1's
		code := ""
		for bitIndex := 0; bitIndex < int(codeLen); bitIndex++ {
			byteIndex := bitIndex / 8
			bitPos := 7 - (bitIndex % 8)
			if (codeBits[byteIndex]>>bitPos)&1 == 1 {
				code += "1"
			} else {
				code += "0"
			}
		}

		// Store the rune and its corresponding code
		codes[rune(r)] = code
	}

	return codes, nil
}

func compressData(input io.Reader, output io.Writer, codes map[rune]string) (uint64, error) {
	var currentByte byte
	var bitCount uint8
	compressedLength := uint64(0)
	buf := make([]byte, 256)

	for {
		n, err := input.Read(buf)
		if err != nil && err != io.EOF {
			return 0, fmt.Errorf(ec.BUFFER_READ_ERROR, err)
		}
		if n == 0 {
			break // EOF reached
		}

		if err := processByte(buf[:n], output, codes, &currentByte, &bitCount, &compressedLength); err != nil {
			return 0, fmt.Errorf(ec.ERROR_COMPRESS, err)
		}
	}
	// if there are remaining bits in the current byte, write them to the output
	if bitCount > 0 {
		// Pad the last byte with zeros
		currentByte <<= 8 - bitCount
		if err := binary.Write(output, binary.LittleEndian, currentByte); err != nil {
			return 0, fmt.Errorf(ec.FILE_WRITE_ERROR, err)
		}
	} else {
		if err := binary.Write(output, binary.LittleEndian, byte(0)); err != nil {
			return 0, fmt.Errorf(ec.FILE_WRITE_ERROR, err)
		}
	}
	// Write the number of bits in the last byte
	if err := binary.Write(output, binary.LittleEndian, bitCount); err != nil {
		return 0, fmt.Errorf(ec.FILE_WRITE_ERROR, err)
	}

	compressedLength += 2 // 1 byte for the last byte and 1 byte for the number of bits in the last byte

	return compressedLength, nil
}

func processByte(buf []byte, output io.Writer, codes map[rune]string, currentByte *byte, bitCount *uint8, compressedLength *uint64) error {

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
					return fmt.Errorf(ec.FILE_WRITE_ERROR, err)
				}
				*currentByte = 0
				*bitCount = 0
				*compressedLength++
			}
		}
	}

	return nil
}

func decompressData(reader io.Reader, writer io.Writer, codes map[rune]string, limiter uint64) error {

	lastByte := make([]byte, 1)
	lastByteCount := make([]byte, 1)

	leftOverByte := uint32(0)
	leftOverByteCount := uint8(0)

	loopFlag := 0
	root := rebuildHuffmanTree(codes)
	currentNode := root

	// if the limiter is not -1, then we only read limiter bytes

	dataRead := uint64(0)

	bytesToRead := BUFFER_SIZE

	for {

		if dataRead <= limiter {
			bytesToRead = int(min(BUFFER_SIZE, limiter-dataRead))
		}

		readBuffer := make([]byte, bytesToRead)
		n, err := reader.Read(readBuffer)
		if err != nil && err != io.EOF {
			return err
		}

		dataRead += uint64(n)

		if n == 0 {

			err := decompressRemainingBits(leftOverByte, leftOverByteCount, uint8(lastByteCount[0]), lastByte[0], writer, root)
			if err != nil {
				return fmt.Errorf(ec.ERROR_COMPRESS, err)
			}

			break
		} else {

			adjustBuffer(loopFlag, &n, &lastByte, &lastByteCount, &readBuffer)

			if err := decompressFullByte(readBuffer, &leftOverByte, &leftOverByteCount, &currentNode, &root, writer); err != nil {
				return fmt.Errorf(ec.ERROR_COMPRESS, err)
			}
		}

		loopFlag = 1
	}

	return nil
}

func adjustBuffer(loopFlag int, n *int, lastByte *[]byte, lastByteCount *[]byte, readBuffer *[]byte) {
	if loopFlag != 0 {
		*n += 2
		//add the last byte and bit count to the readBuffer
		leftOverData := append(*lastByte, *lastByteCount...)
		*readBuffer = append(leftOverData, *readBuffer...)
	}

	// add the last byte from the chunk to the lastByte buffer
	*lastByte = (*readBuffer)[*n-2 : *n-1]
	// add the last byte's bit count to the lastByteCount buffer
	*lastByteCount = (*readBuffer)[*n-1 : *n]
	// remove last 2 bytes from the chunk
	*readBuffer = (*readBuffer)[:*n-2]
}

func decompressFullByte(readBuffer []byte, leftOverByte *uint32, leftOverByteCount *uint8, currentNode **Node, root **Node, writer io.Writer) error {
	for _, b := range readBuffer {
		for i := 7; i >= 0; i-- {
			bit := (b >> i) & 1
			*leftOverByte <<= 1
			*leftOverByte |= uint32(bit)
			*leftOverByteCount++
			if bit == 0 {
				*currentNode = (*currentNode).left
			} else {
				*currentNode = (*currentNode).right
			}
			if (*currentNode).left == nil && (*currentNode).right == nil {
				if _, err := writer.Write([]byte{byte((*currentNode).char)}); err != nil {
					return fmt.Errorf(ec.FILE_WRITE_ERROR, err)
				}
				*currentNode = *root
				*leftOverByte = 0
				*leftOverByteCount = 0
			}
		}
	}

	return nil
}

func decompressRemainingBits(remainingBits uint32, remainingBitsLen uint8, numOfBits uint8, lastByte byte, output io.Writer, root *Node) error {
	// include the last byte in the remaining bits
	for j := 7; j >= 8-int(numOfBits); j-- {
		bit := (lastByte >> j) & 1
		//remainingBits = (remainingBits << 1) | bit
		remainingBits <<= 1
		remainingBits |= uint32(bit)
		remainingBitsLen++
	}

	currentNode := root

	for i := int(remainingBitsLen - 1); i >= 0; i-- {

		bit := (remainingBits >> i) & 1

		if bit == 0 {
			currentNode = currentNode.left
		} else {
			currentNode = currentNode.right
		}

		if currentNode.left == nil && currentNode.right == nil {
			if _, err := output.Write([]byte{byte(currentNode.char)}); err != nil {
				return fmt.Errorf(ec.BUFFER_WRITE_ERROR, err)
			}
			currentNode = root
		}
	}

	return nil
}

// Zip compresses data using Huffman coding and writes the compressed data to the output stream.
func Zip(files []utils.FileData, output io.Writer) error {

	codes, err := generateCodes(&files, output)
	if err != nil {
		return fmt.Errorf("error preparing codes: %w", err)
	}

	// Write the number of files
	if err := writeNumOfFiles(uint64(len(files)), output); err != nil {
		return fmt.Errorf(ec.FILE_WRITE_ERROR, err)
	}

	for _, file := range files {
		reader := file.Reader

		//Compress and write the file name
		if err = writeFileName(file.Name, output, codes); err != nil {
			return fmt.Errorf(ec.FILE_WRITE_ERROR, err)
		}

		//write 32 bit 0 for the compressed size
		if err := binary.Write(output, binary.LittleEndian, uint64(0)); err != nil {
			return fmt.Errorf(ec.FILE_WRITE_ERROR, err)
		}
		//Compress and write the data
		compressedLen, err := compressData(reader, output, codes)

		if err != nil {
			return fmt.Errorf(ec.ERROR_COMPRESS, err)
		}

		//seek back to compressedLen bytes and write the compressed size
		if _, err := output.(io.Seeker).Seek(-int64(compressedLen+8), io.SeekCurrent); err != nil { // +4 for the 4 bytes of compressed size (uint32 -> 4 bytes) | 8bit = 1byte, 32bit = 4byte
			return fmt.Errorf("error seeking back to write the compressed size: %w", err)
		}

		if err := binary.Write(output, binary.LittleEndian, compressedLen); err != nil {
			return fmt.Errorf(ec.FILE_WRITE_ERROR, err)
		}

		//seek back to the end of the file
		if _, err := output.(io.Seeker).Seek(0, io.SeekEnd); err != nil {
			return fmt.Errorf("error seeking to the end of the file: %w", err)
		}
	}

	return nil
}

func generateCodes(files *[]utils.FileData, output io.Writer) (map[rune]string, error) {
	freq := make(map[rune]int)
	//first, we need to get the frequency map
	for _, file := range *files {

		//Get frequency map of the input file name
		nameBuf := bytes.NewReader([]byte(file.Name))
		if err := getFrequencyMap(nameBuf, &freq); err != nil {
			return nil, fmt.Errorf("error generating frequency map for filename: %w", err)
		}

		//Get frequency map of the input data
		if err := getFrequencyMap(file.Reader, &freq); err != nil {
			return nil, fmt.Errorf("error generating frequency map for filedata: %w", err)
		}

		//reset the seek
		if _, err := file.Reader.(io.Seeker).Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
	}

	// Build Huffman codes
	codes, err := GetHuffmanCodes(&freq)
	if err != nil {
		return nil, err
	}

	// Write frequency map and Huffman codes to the output
	if err := WriteHuffmanCodes(output, codes); err != nil {
		return nil, fmt.Errorf(ec.FAILED_WRITE_HUFFMAN_CODES, err)
	}

	return codes, nil
}

func writeFileName(fileName string, output io.Writer, codes map[rune]string) error {
	// Write the file name length after compressing it
	nameBuf := bytes.NewReader([]byte(fileName))

	compressedNameBuf := bytes.NewBuffer([]byte{})

	compLen, err := compressData(nameBuf, compressedNameBuf, codes)
	if err != nil {
		return fmt.Errorf(ec.ERROR_COMPRESS, err)
	}

	// write length of the file name buffer
	if err := binary.Write(output, binary.LittleEndian, uint16(compLen)); err != nil {
		return fmt.Errorf(ec.FILE_WRITE_ERROR, err)
	}

	// write the compressed file name
	if err := binary.Write(output, binary.LittleEndian, compressedNameBuf.Bytes()); err != nil {
		return fmt.Errorf(ec.FILE_WRITE_ERROR, err)
	}

	return nil
}

func readNumOfFiles(input io.Reader) (uint64, error) {
	var numOfFiles uint64
	if err := binary.Read(input, binary.LittleEndian, &numOfFiles); err != nil {
		return 0, fmt.Errorf(ec.FILE_READ_ERROR, err)
	}

	return numOfFiles, nil
}

func writeNumOfFiles(numOfFiles uint64, output io.Writer) error {

	if err := binary.Write(output, binary.LittleEndian, numOfFiles); err != nil {
		return fmt.Errorf(ec.FILE_WRITE_ERROR, err)
	}

	return nil
}

func readFileName(input io.Reader, codes map[rune]string) (string, error) {

	var nameLen uint16
	if err := binary.Read(input, binary.LittleEndian, &nameLen); err != nil {
		return "", fmt.Errorf(ec.FILE_READ_ERROR, err)
	}

	buf := make([]byte, nameLen)
	if err := binary.Read(input, binary.LittleEndian, buf); err != nil {
		return "", fmt.Errorf(ec.FILE_READ_ERROR, err)
	}

	compressedFilename := bytes.NewBuffer(buf)

	nameBuffer := bytes.NewBuffer([]byte{})
	if err := decompressData(compressedFilename, nameBuffer, codes, uint64(nameLen)); err != nil {
		return "", fmt.Errorf(ec.ERROR_DECOMPRESS, err)
	}

	name := nameBuffer.String()

	return name, nil
}

// Unzip decompresses data using Huffman coding and writes the decompressed data to the output stream.
func Unzip(input io.Reader, outputPath string) ([]string, error) {

	if outputPath == "" {
		outputPath = "." // Use the current directory if no output path is provided
	}

	codes, err := ReadHuffmanCodes(input)
	if err != nil {
		return nil, fmt.Errorf(ec.FAILED_READ_HUFFMAN_CODES, err)
	}

	numOfFiles, err := readNumOfFiles(input)
	if err != nil {
		return nil, err
	}

	if numOfFiles < 1 {
		return nil, errors.New("no files to decompress")
	}

	filePaths := []string{}

	for i := uint64(0); i < numOfFiles; i++ {
		// get the file name
		fileName, err := readFileName(input, codes)
		if err != nil {
			return nil, err
		}

		fileName = path.Join(outputPath, fileName)

		dir := path.Dir(fileName)

		if err := utils.MakeOutputDir(dir); err != nil {
			return nil, fmt.Errorf(ec.ERROR_CREATE_DIR, err)
		}

		// writer
		outputFile, err := os.Create(fileName)
		if err != nil {
			return nil, fmt.Errorf(ec.FILE_CREATE_ERROR, err)
		}

		// read the compressed size
		var compressedSize uint64
		if err := binary.Read(input, binary.LittleEndian, &compressedSize); err != nil {
			return nil, fmt.Errorf(ec.FILE_READ_ERROR, err)
		}

		err = decompressData(input, outputFile, codes, compressedSize)
		if err != nil {
			return nil, fmt.Errorf(ec.ERROR_DECOMPRESS, err)
		}

		outputFile.Close()

		filePaths = append(filePaths, fileName)
	}

	return filePaths, nil
}
