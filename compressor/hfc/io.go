package hfc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"file-compressor/constants"
	"file-compressor/utils"
)


// getFrequencyMap reads data from the provided io.Reader and populates the given frequency map
// with the count of each rune encountered in the input.
//
// Parameters:
//   - input: an io.Reader from which data is read.
//   - freq: a pointer to a map where the frequency of each rune will be stored.
//
// Returns:
//   - error: an error if reading from the input fails, otherwise nil.
func getFrequencyMap(input io.Reader, freq *map[rune]int) error {
	buf := make([]byte, constants.BUFFER_SIZE)
	for {
		n, err := input.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf(constants.BUFFER_READ_ERROR, err)
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

// WriteHuffmanCodes writes Huffman codes to the provided io.Writer.
// It first writes the number of codes, followed by each rune and its corresponding code.
// Each code is written as a sequence of bytes, with the length of the code being a multiple of 8 bits.
// If the length of the code is not a multiple of 8, it is rounded up to the nearest higher multiple of 8.
//
// Parameters:
//   - file: An io.Writer where the Huffman codes will be written.
//   - codes: A map where the keys are runes and the values are their corresponding Huffman codes as strings.
//
// Returns:
//   - error: An error if any occurs during writing, otherwise nil.
func WriteHuffmanCodes(file io.Writer, codes map[rune]string) error {
	//write the number of codes
	binary.Write(file, binary.LittleEndian, uint64(len(codes)))

	for k, v := range codes {
		//write the rune
		binary.Write(file, binary.LittleEndian, uint32(k))
		//write the length of the code
		// codelen must be a multiple of 8, if not round it up to the nearest higher multiple of 8
		codeLen := len(v)
		binary.Write(file, binary.LittleEndian, uint8(codeLen)) // bit length of the code
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

// ReadHuffmanCodes reads Huffman codes from the provided io.Reader and returns a map
// where the keys are runes and the values are their corresponding Huffman codes as strings.
// The function expects the input to be in a specific binary format:
// - The first 4 bytes represent the number of Huffman codes (uint32).
// - For each Huffman code:
//   - The next 4 bytes represent the rune (uint32).
//   - The next byte represents the length of the Huffman code (uint8).
//   - The subsequent bytes contain the Huffman code bits, packed into bytes.
//
// Parameters:
// - file: An io.Reader from which the Huffman codes will be read.
//
// Returns:
// - A map[rune]string where each rune is mapped to its corresponding Huffman code.
// - An error if there is an issue reading from the file or if the data is in an unexpected format.
func ReadHuffmanCodes(file io.Reader) (map[rune]string, error) {
	codes := make(map[rune]string)

	// Read the number of codes (4 bytes, uint32)
	var numCodes uint64
	if err := binary.Read(file, binary.LittleEndian, &numCodes); err != nil {
		return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
	}

	// Read each code
	for i := uint64(0); i < numCodes; i++ {
		// Read the rune
		var r uint32
		if err := binary.Read(file, binary.LittleEndian, &r); err != nil {
			return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
		}

		// Read the length of the Huffman code (1 byte)
		var codeLen uint8
		if err := binary.Read(file, binary.LittleEndian, &codeLen); err != nil {
			return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
		}

		// Read the code bits (packed into bytes)
		codeBits := make([]byte, int((codeLen+7)/8)) // (codeLen+7)/8 ensures enough space to hold all bits
		if _, err := file.Read(codeBits); err != nil {
			return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
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

// compressData compresses data from the input reader and writes the compressed data to the output writer
// using the provided Huffman codes.
//
// Parameters:
//   - input: An io.Reader from which the data to be compressed is read.
//   - output: An io.Writer to which the compressed data is written.
//   - codes: A map of runes to their corresponding Huffman codes.
//
// Returns:
//   - uint64: The length of the compressed data in bytes.
//   - error: An error if any occurs during the compression process.
//
// The function reads data from the input in chunks, processes each chunk to compress it using the provided
// Huffman codes, and writes the compressed data to the output. It handles padding of the last byte and writes
// the number of bits used in the last byte to the output. If an error occurs during reading, processing, or
// writing, the function returns the error.
func compressData(input io.Reader, output io.Writer, codes map[rune]string) (uint64, error) {
	var currentByte byte
	var bitCount uint8
	compressedLength := uint64(0)
	buf := make([]byte, constants.BUFFER_SIZE)

	for {
		n, err := input.Read(buf)
		if err != nil && err != io.EOF {
			return 0, fmt.Errorf(constants.BUFFER_READ_ERROR, err)
		}
		if n == 0 {
			break // EOF reached
		}

		if err := processByte(buf[:n], output, codes, &currentByte, &bitCount, &compressedLength); err != nil {
			return 0, fmt.Errorf(constants.ERROR_COMPRESS, err)
		}
	}
	// if there are remaining bits in the current byte, write them to the output
	if bitCount > 0 {
		// Pad the last byte with zeros
		currentByte <<= 8 - bitCount
		if err := binary.Write(output, binary.LittleEndian, currentByte); err != nil {
			return 0, fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}
	} else {
		if err := binary.Write(output, binary.LittleEndian, byte(0)); err != nil {
			return 0, fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}
	}
	// Write the number of bits in the last byte
	if err := binary.Write(output, binary.LittleEndian, bitCount); err != nil {
		return 0, fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	compressedLength += 2 // 1 byte for the last byte and 1 byte for the number of bits in the last byte

	return compressedLength, nil
}

// processByte processes a buffer of bytes, compressing it using Huffman codes and writing the result to an output writer.
// 
// Parameters:
//   - buf: A slice of bytes to be processed.
//   - output: An io.Writer where the compressed data will be written.
//   - codes: A map of runes to their corresponding Huffman codes as strings.
//   - currentByte: A pointer to the current byte being constructed from the Huffman codes.
//   - bitCount: A pointer to the count of bits currently in the current byte.
//   - compressedLength: A pointer to the total length of the compressed data.
//
// Returns:
//   - error: An error if there is no Huffman code for a character in the buffer or if there is an issue writing to the output.
//
// The function iterates over each byte in the buffer, looks up its corresponding Huffman code, and writes the bits of the code
// to the current byte. When the current byte is full (i.e., 8 bits), it writes the byte to the output and resets the current byte
// and bit count. The compressed length is incremented each time a byte is written to the output.
func processByte(buf []byte, output io.Writer, codes map[rune]string, currentByte *byte, bitCount *uint8, compressedLength *uint64) error {

	for _, b := range buf {
		char := rune(b)
		code, exists := codes[char]
		if !exists {
			return fmt.Errorf("no Huffman code for character (%b) - (%c)", char, char)
			//continue
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
					return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
				}
				*currentByte = 0
				*bitCount = 0
				*compressedLength++
			}
		}
	}

	return nil
}

// decompressData decompresses data from the provided reader and writes the decompressed data to the provided writer.
// It uses the provided Huffman codes to decode the data and respects the limiter for the maximum number of bytes to read.
//
// Parameters:
//   - reader: An io.Reader from which compressed data is read.
//   - writer: An io.Writer to which decompressed data is written.
//   - codes: A map of Huffman codes used for decompression.
//   - limiter: A uint64 value specifying the maximum number of bytes to read. If set to -1, there is no limit.
//
// Returns:
//   - error: An error if decompression fails, otherwise nil.
func decompressData(reader io.Reader, writer io.Writer, codes map[rune]string, limiter uint64) error {

	lastByte := make([]byte, 1) // last byte read from the reader
	lastByteCount := make([]byte, 1) // number of bits in the last byte

	leftOverByte := uint32(0)
	leftOverByteCount := uint8(0)

	loopFlag := 0
	root := rebuildHuffmanTree(codes)
	currentNode := root

	// if the limiter is not -1, then we only read limiter bytes

	dataRead := uint64(0)

	bytesToRead := constants.BUFFER_SIZE

	for {

		if dataRead <= limiter {
			bytesToRead = int(min(constants.BUFFER_SIZE, limiter-dataRead))
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
				return fmt.Errorf(constants.ERROR_COMPRESS, err)
			}

			break
		} else {

			adjustBuffer(loopFlag, &n, &lastByte, &lastByteCount, &readBuffer)

			if err := decompressFullByte(readBuffer, &leftOverByte, &leftOverByteCount, &currentNode, &root, writer); err != nil {
				return fmt.Errorf(constants.ERROR_COMPRESS, err)
			}
		}

		loopFlag = 1
	}

	return nil
}

// adjustBuffer modifies the read buffer based on the loop flag and updates the last byte and its count.
//
// Parameters:
// - loopFlag: An integer flag that determines if the buffer should be adjusted.
// - n: A pointer to an integer that tracks the current position in the buffer.
// - lastByte: A pointer to a byte slice that stores the last byte from the previous chunk.
// - lastByteCount: A pointer to a byte slice that stores the bit count of the last byte from the previous chunk.
// - readBuffer: A pointer to a byte slice that represents the current read buffer.
//
// If loopFlag is non-zero, the function increments the position by 2, appends the last byte and its count to the 
// beginning of the read buffer, and then updates the last byte and its count based on the current position. 
// Finally, it removes the last 2 bytes from the read buffer.
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

// decompressFullByte processes a buffer of bytes to decompress data using a Huffman tree.
// It traverses the tree based on the bits of each byte in the buffer and writes the decompressed
// characters to the provided writer.
//
// Parameters:
// - readBuffer: A slice of bytes containing the compressed data.
// - leftOverByte: A pointer to an uint32 that holds the leftover bits from the previous byte.
// - leftOverByteCount: A pointer to an uint8 that counts the number of leftover bits.
// - currentNode: A double pointer to the current node in the Huffman tree.
// - root: A double pointer to the root node of the Huffman tree.
// - writer: An io.Writer where the decompressed data will be written.
//
// Returns:
// - error: An error if writing to the writer fails, otherwise nil.
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
					return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
				}
				*currentNode = *root
				*leftOverByte = 0
				*leftOverByteCount = 0
			}
		}
	}

	return nil
}

// decompressRemainingBits processes the remaining bits from the last byte of a compressed stream,
// traversing the Huffman tree to decode the bits and write the corresponding characters to the output.
//
// Parameters:
//   - remainingBits: The bits that are left to be processed.
//   - remainingBitsLen: The length of the remaining bits.
//   - numOfBits: The number of bits to process from the last byte.
//   - lastByte: The last byte from the compressed stream.
//   - output: The writer where the decompressed data will be written.
//   - root: The root node of the Huffman tree.
//
// Returns:
//   - error: An error if writing to the output fails, otherwise nil.
func decompressRemainingBits(remainingBits uint32, remainingBitsLen uint8, numOfBits uint8, lastByte byte, output io.Writer, root *Node) error {
	// include the last byte in the remaining bits
	for j := 7; j >= 8-int(numOfBits); j-- {
		bit := (lastByte >> j) & 1
		//remainingBits = (remainingBits << 1) | bit
		remainingBits <<= 1
		remainingBits |= uint32(bit)
		remainingBitsLen++
	}

	if remainingBitsLen == 0 {
		return nil
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
				return fmt.Errorf(constants.BUFFER_WRITE_ERROR, err)
			}
			currentNode = root
		}
	}

	return nil
}

// Zip compresses data using Huffman coding and writes the compressed data to the output stream.
// Zip compresses multiple files and writes the compressed data to the provided output writer.
// It first generates compression codes for the files, writes the number of files, and then processes each file individually.
// For each file, it compresses and writes the file name, writes a placeholder for the compressed size, compresses the file data,
// and then updates the placeholder with the actual compressed size.
//
// Parameters:
//   - files: A slice of utils.FileData representing the files to be compressed.
//   - output: An io.Writer where the compressed data will be written.
//
// Returns:
//   - error: An error if any step in the compression process fails.
func Zip(files []utils.FileData, output io.Writer) error {

	codes, err := generateCodes(&files, output)
	if err != nil {
		return fmt.Errorf("error preparing codes: %w", err)
	}

	// Write the number of files
	if err := writeNumOfFiles(uint64(len(files)), output); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	for _, file := range files {
		reader := file.Reader

		//Compress and write the file name
		if err = writeFileName(file.Name, output, codes); err != nil {
			return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}

		//write 64 bit 0 for the compressed size
		if err := binary.Write(output, binary.LittleEndian, uint64(0)); err != nil {
			return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}
		//Compress and write the data
		compressedLen, err := compressData(reader, output, codes)

		if err != nil {
			return fmt.Errorf(constants.ERROR_COMPRESS, err)
		}

		//seek back to compressedLen bytes and write the compressed size
		if _, err := output.(io.Seeker).Seek(-int64(compressedLen+8), io.SeekCurrent); err != nil { // +4 for the 4 bytes of compressed size (uint64 -> 8 bytes) | 8bit = 1byte, 64bit = 8byte
			return fmt.Errorf("error seeking back to write the compressed size: %w", err)
		}

		if err := binary.Write(output, binary.LittleEndian, compressedLen); err != nil {
			return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}

		//seek back to the end of the file
		if _, err := output.(io.Seeker).Seek(0, io.SeekEnd); err != nil {
			return fmt.Errorf("error seeking to the end of the file: %w", err)
		}
	}

	return nil
}

// generateCodes generates Huffman codes for the given files and writes the frequency map and codes to the output.
// It returns a map of runes to their corresponding Huffman codes.
//
// Parameters:
// - files: A pointer to a slice of utils.FileData, where each FileData contains the file name and a reader for the file content.
// - output: An io.Writer where the frequency map and Huffman codes will be written.
//
// Returns:
// - A map[rune]string representing the Huffman codes for each rune.
// - An error if there is any issue during the process of generating the frequency map, building Huffman codes, or writing the codes to the output.
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
		return nil, fmt.Errorf(constants.FAILED_WRITE_HUFFMAN_CODES, err)
	}

	return codes, nil
}

func writeFileName(fileName string, output io.Writer, codes map[rune]string) error {
	// Write the file name length after compressing it
	nameBuf := bytes.NewReader([]byte(fileName))

	compressedNameBuf := bytes.NewBuffer([]byte{})

	compLen, err := compressData(nameBuf, compressedNameBuf, codes)
	if err != nil {
		return fmt.Errorf(constants.ERROR_COMPRESS, err)
	}
	// write length of the file name buffer
	if err := binary.Write(output, binary.LittleEndian, uint16(compLen)); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	// write the compressed file name
	if err := binary.Write(output, binary.LittleEndian, compressedNameBuf.Bytes()); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	return nil
}

func readNumOfFiles(input io.Reader) (uint64, error) {
	var numOfFiles uint64
	if err := binary.Read(input, binary.LittleEndian, &numOfFiles); err != nil {
		return 0, fmt.Errorf(constants.FILE_READ_ERROR, err)
	}

	return numOfFiles, nil
}

func writeNumOfFiles(numOfFiles uint64, output io.Writer) error {

	if err := binary.Write(output, binary.LittleEndian, numOfFiles); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	return nil
}

func readFileName(input io.Reader, codes map[rune]string) (string, error) {

	var nameLen uint16
	if err := binary.Read(input, binary.LittleEndian, &nameLen); err != nil {
		return "", fmt.Errorf(constants.FILE_READ_ERROR, err)
	}

	buf := make([]byte, nameLen)
	if err := binary.Read(input, binary.LittleEndian, buf); err != nil {
		return "", fmt.Errorf(constants.FILE_READ_ERROR, err)
	}

	compressedFilename := bytes.NewBuffer(buf)

	nameBuffer := bytes.NewBuffer([]byte{})
	if err := decompressData(compressedFilename, nameBuffer, codes, uint64(nameLen)); err != nil {
		return "", fmt.Errorf(constants.ERROR_DECOMPRESS, err)
	}

	name := nameBuffer.String()

	return name, nil
}


// Unzip decompresses data from the provided io.Reader and writes the decompressed files to the specified output path.
// If the output path is an empty string, the current directory is used.
//
// Parameters:
//   - input: An io.Reader from which the compressed data is read.
//   - outputPath: A string specifying the directory where the decompressed files will be written.
//
// Returns:
//   - A slice of strings containing the paths of the decompressed files.
//   - An error if any issue occurs during the decompression process.
//
// The function performs the following steps:
//   1. Reads Huffman codes from the input.
//   2. Reads the number of files to be decompressed.
//   3. Iterates over each file, reading its name and creating the necessary directories.
//   4. Creates the output file and reads its compressed size.
//   5. Decompresses the data and writes it to the output file.
//   6. Closes the output file and appends its path to the result slice.
//
// Possible errors include issues with reading Huffman codes, reading the number of files, creating directories, 
// creating output files, reading compressed sizes, and decompressing data.
func Unzip(input io.Reader, outputPath string) ([]string, error) {

	if outputPath == "" {
		outputPath = "." // Use the current directory if no output path is provided
	}

	codes, err := ReadHuffmanCodes(input)
	if err != nil {
		return nil, fmt.Errorf(constants.FAILED_READ_HUFFMAN_CODES, err)
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

		fileName = filepath.Join(outputPath, fileName)

		dir := filepath.Dir(fileName)

		if err := utils.MakeOutputDir(dir); err != nil {
			return nil, fmt.Errorf(constants.ERROR_CREATE_DIR, err)
		}

		// writer
		outputFile, err := os.Create(fileName)
		if err != nil {
			return nil, fmt.Errorf(constants.FILE_CREATE_ERROR, err)
		}

		// read the compressed size
		var compressedSize uint64
		if err := binary.Read(input, binary.LittleEndian, &compressedSize); err != nil {
			return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
		}

		err = decompressData(input, outputFile, codes, compressedSize)
		if err != nil {
			return nil, fmt.Errorf(constants.ERROR_DECOMPRESS, err)
		}

		outputFile.Close()

		filePaths = append(filePaths, fileName)
	}

	return filePaths, nil
}
