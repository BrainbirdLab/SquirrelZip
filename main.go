package main

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

type File struct {
	Name    string
	Content []byte
}

// Huffman coding structures

type Node struct {
	char  rune
	freq  int
	left  *Node
	right *Node
}

type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].freq < pq[j].freq
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*Node))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0:n-1]
	return item
}

func buildHuffmanTree(freq map[rune]int) *Node {
	pq := make(PriorityQueue, len(freq))
	i := 0
	for char, f := range freq {
		pq[i] = &Node{char: char, freq: f}
		i++
	}
	heap.Init(&pq)

	for len(pq) > 1 {
		left := heap.Pop(&pq).(*Node)
		right := heap.Pop(&pq).(*Node)

		internal := &Node{
			char:  '\x00', // internal node character
			freq:  left.freq + right.freq,
			left:  left,
			right: right,
		}
		heap.Push(&pq, internal)
	}

	if len(pq) > 0 {
		return heap.Pop(&pq).(*Node) // root of Huffman tree
	}
	return nil
}

func buildHuffmanCodes(root *Node) map[rune]string {
	codes := make(map[rune]string)
	var build func(node *Node, code string)
	build = func(node *Node, code string) {
		if node == nil {
			return
		}
		if node.left == nil && node.right == nil {
			codes[node.char] = code
			return
		}
		build(node.left, code+"0")
		build(node.right, code+"1")
	}
	build(root, "")
	return codes
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

func Zip(files []File) File {
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

	return File{
		Name:    "compressed.bin",
		Content: buf.Bytes(),
	}
}

func Unzip(file File) []File {
	var files []File

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
		files = append(files, File{
			Name:    string(name),
			Content: content,
		})
	}

	return files
}

func rebuildHuffmanTree(codes map[rune]string) *Node {
	var root *Node
	for char, code := range codes {
		if root == nil {
			root = &Node{}
		}
		current := root
		for _, bit := range code {
			if bit == '0' {
				if current.left == nil {
					current.left = &Node{}
				}
				current = current.left
			} else if bit == '1' {
				if current.right == nil {
					current.right = &Node{}
				}
				current = current.right
			}
		}
		current.char = char
	}
	return root
}

func Compress(filenameStrs []string, outputDir *string) {
	if len(filenameStrs) == 0 {
		fmt.Println("No files to compress.")
		return
	}

	if *outputDir == "" {
		*outputDir = path.Dir(filenameStrs[0])
	}

	var files []File

	// Read files
	for _, filename := range filenameStrs {
		// Read file content and append to files
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			fmt.Println("File does not exist.")
			return
		}

		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println(err)
			return
		}

		files = append(files, File{
			Name:   path.Base(filename),
			Content: content,
		})
	}

	// Compress files
	compressedFile := Zip(files)

	// Write compressed file in the current directory + /compressed directory
	os.WriteFile(*outputDir + "/" + compressedFile.Name, compressedFile.Content, 0644)
}

func Decompress(filenameStrs []string, outputDir *string) {
	if len(filenameStrs) == 0 {
		fmt.Println("No files to decompress.")
		return
	}

	if *outputDir == "" {
		*outputDir = path.Dir(filenameStrs[0])
	}

	// Read compressed file
	if _, err := os.Stat(filenameStrs[0]); os.IsNotExist(err) {
		fmt.Println("File does not exist.")
		return
	}

	content, err := os.ReadFile(filenameStrs[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	// Unzip file
	files := Unzip(File{
		Name:    path.Base(filenameStrs[0]),
		Content: content,
	})

	// Write decompressed files
	for _, file := range files {
		fmt.Printf("Decompressed file: %s\n", file.Name)
		os.WriteFile(*outputDir + "/" + file.Name, file.Content, 0644)
	}
}

func main() {
	//test files path '/test'

	//cli arguments
	inputFiles := flag.String("i", "", "Input files to be compressed")
	outputDir := flag.String("o", "", "Output directory for compressed files (Optional)")
	readAllFiles := flag.Bool("a", false, "Read all files in the test directory")
	decompressMode := flag.Bool("d", false, "Decompress mode")
	flag.Parse()

	if *inputFiles == "" {
		fmt.Println("Input files are required.")
		flag.Usage()
		os.Exit(1)
	}

	// Read all files
	var filenameStrs []string

	// Split input files by comma and trim spaces and quotes(if any, `'` | `"`)

	if *readAllFiles {

		fmt.Printf("Reading all files in the directory: %s\n", *inputFiles)

		//filenameStrs is the folder name
		dir := inputFiles
		//if dir exists
		if _, err := os.Stat(*dir); os.IsNotExist(err) {
			fmt.Println("Directory does not exist.")
			return
		}

		//read all filenames in the directory
		files, err := os.ReadDir(*dir)
		if err != nil {
			fmt.Println("Error reading directory files. Please check the directory path.")
			return
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			filenameStrs = append(filenameStrs, *dir + "/" + file.Name())
		}
	} else {
		for _, filename := range strings.Split(*inputFiles, ",") {
			filenameStrs = append(filenameStrs, strings.Trim(filename, " '\""))
		}
	}

	fmt.Printf("Files to compress/decompress: %v\n", filenameStrs)

	if *decompressMode {
		Decompress(filenameStrs, outputDir)
	} else {
		Compress(filenameStrs, outputDir)
	}
}
