package compressor

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path"

	"file-compressor/utils"
)

func Compress(filenameStrs []string, outputDir *string, password *string) error {

	if len(filenameStrs) == 0 {
		return errors.New("no files to compress")
	}

	if *outputDir == "" {
		*outputDir = path.Dir(filenameStrs[0])
	}

	var files []utils.File

	// Read files
	for _, filename := range filenameStrs {
		// Read file content and append to files
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return errors.New("file does not exist")
		}

		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		files = append(files, utils.File{
			Name:   path.Base(filename),
			Content: content,
		})
	}

	// Compress files
	compressedFile := Zip(files)

	var err error

	compressedFile.Content, err = utils.Encrypt(compressedFile.Content, *password)

	if err != nil {
		return err
	}

	// Write compressed file in the current directory + /compressed directory
	os.WriteFile(*outputDir + "/" + compressedFile.Name, compressedFile.Content, 0644)

	return nil
}

func Decompress(filenameStrs []string, outputDir *string, password *string) error {
	if len(filenameStrs) == 0 {
		return errors.New("no files to decompress")
	}

	if *outputDir == "" {
		*outputDir = path.Dir(filenameStrs[0])
	}

	// Read compressed file
	if _, err := os.Stat(filenameStrs[0]); os.IsNotExist(err) {
		return errors.New("file does not exist")
	}

	content, err := os.ReadFile(filenameStrs[0])
	if err != nil {
		return err
	}

	content, err = utils.Decrypt(content, *password)
	if err != nil {
		return err
	}

	// Unzip file
	files := Unzip(utils.File{
		Name:    path.Base(filenameStrs[0]),
		Content: content,
	})

	// Write decompressed files
	for _, file := range files {
		utils.ColorPrint(utils.GREEN, fmt.Sprintf("Decompressed file: %s\n", file.Name))
		os.WriteFile(*outputDir + "/" + file.Name, file.Content, 0644)
	}

	return nil
}


func Zip(files []utils.File) utils.File {
	var buf bytes.Buffer

	// Write header count of files
	binary.Write(&buf, binary.BigEndian, uint32(len(files)))

	// Create raw content buffer
	var rawContent bytes.Buffer
	for _, file := range files {
		// Write filename length and filename
		filenameLen := uint32(len(file.Name))
		binary.Write(&rawContent, binary.BigEndian, filenameLen)
		rawContent.WriteString(file.Name)

		// Write content length and content
		contentLen := uint32(len(file.Content))
		binary.Write(&rawContent, binary.BigEndian, contentLen)
		rawContent.Write(file.Content)
	}

	// Compress rawContent using Huffman coding
	freq := make(map[rune]int)
	for _, b := range rawContent.Bytes() {
		freq[rune(b)]++
	}
	root := buildHuffmanTree(freq)
	codes := buildHuffmanCodes(root)
	compressedContent := compressData(rawContent.Bytes(), codes)

	// Write compressed content length to buffer
	binary.Write(&buf, binary.BigEndian, uint32(len(compressedContent)))
	// Write compressed content to buffer
	buf.Write(compressedContent)

	// Write Huffman codes length to buffer
	binary.Write(&buf, binary.BigEndian, uint32(len(codes)))
	// Write Huffman codes to buffer
	for char, code := range codes {
		binary.Write(&buf, binary.BigEndian, char)
		binary.Write(&buf, binary.BigEndian, uint32(len(code)))
		buf.WriteString(code)
	}

	return utils.File{
		Name:   	"compressed.bin",
		Content: 	buf.Bytes(),
	}
}

func Unzip(file utils.File) []utils.File {
	var files []utils.File

	// Read file content
	buf := bytes.NewBuffer(file.Content)

	// Read number of files in header
	var numFiles uint32
	binary.Read(buf, binary.BigEndian, &numFiles)

	// Read compressed content length
	var compressedContentLength uint32
	binary.Read(buf, binary.BigEndian, &compressedContentLength)
	compressedContent := make([]byte, compressedContentLength)
	buf.Read(compressedContent)

	// Read Huffman codes length
	var codesLength uint32
	binary.Read(buf, binary.BigEndian, &codesLength)

	// Read Huffman codes
	codes := make(map[rune]string)
	for i := uint32(0); i < codesLength; i++ {
		var char rune
		binary.Read(buf, binary.BigEndian, &char)
		var codeLength uint32
		binary.Read(buf, binary.BigEndian, &codeLength)
		code := make([]byte, codeLength)
		buf.Read(code)
		codes[char] = string(code)
	}

	// Rebuild Huffman tree using codes
	var root *Node
	if len(codes) > 0 {
		root = rebuildHuffmanTree(codes)
	}

	decompressedContent := decompressData(compressedContent, root)

	// Parse decompressed content to extract files
	decompressedContentBuf := bytes.NewBuffer(decompressedContent)
	for f := uint32(0); f < numFiles; f++ {
		// Read filename length
		var nameLength uint32
		binary.Read(decompressedContentBuf, binary.BigEndian, &nameLength)
		// Read filename
		name := make([]byte, nameLength)
		decompressedContentBuf.Read(name)
		// Read content length
		var contentLength uint32
		binary.Read(decompressedContentBuf, binary.BigEndian, &contentLength)
		// Read content
		content := make([]byte, contentLength)
		decompressedContentBuf.Read(content)
		files = append(files, utils.File{
			Name:    string(name),
			Content: content,
		})
	}

	return files
}


func compressData(data []byte, codes map[rune]string) []byte {
	var buf bytes.Buffer
	var bitBuffer uint64
	var bitLength uint
	for _, b := range data {
		code := codes[rune(b)]
		for _, bit := range code {
			bitBuffer <<= 1
			bitLength++
			if bit == '1' {
				bitBuffer |= 1
			}
			if bitLength == 64 {
				binary.Write(&buf, binary.BigEndian, bitBuffer)
				bitBuffer = 0
				bitLength = 0
			}
		}
	}
	if bitLength > 0 {
		bitBuffer <<= (64 - bitLength)
		binary.Write(&buf, binary.BigEndian, bitBuffer)
	}
	return buf.Bytes()
}

func decompressData(data []byte, root *Node) []byte {
	var buf bytes.Buffer
	if root == nil {
		return buf.Bytes()
	}

	node := root
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			bit := (b >> i) & 1
			if bit == 0 {
				node = node.left
			} else {
				node = node.right
			}
			if node.left == nil && node.right == nil {
				buf.WriteByte(byte(node.char))
				node = root
			}
		}
	}
	return buf.Bytes()
}