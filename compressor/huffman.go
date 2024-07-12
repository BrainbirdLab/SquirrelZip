package compressor

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"file-compressor/utils"
	"fmt"
)

// Node represents a node in the Huffman tree
type Node struct {
	char  rune  // Character stored in the node
	freq  int   // Frequency of the character
	left  *Node // Left child node
	right *Node // Right child node
}

// PriorityQueue implements heap.Interface and holds Nodes
type PriorityQueue []*Node

// Len returns the number of items in the priority queue
func (pq PriorityQueue) Len() int { return len(pq) }

// Less defines the ordering of items in the priority queue based on frequency
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].freq < pq[j].freq
}

// Swap swaps two items in the priority queue
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

// Push adds an item (Node) to the priority queue
func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*Node))
}

// Pop removes and returns the highest priority item (Node) from the priority queue
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// buildHuffmanTree builds the Huffman tree from character frequencies
func buildHuffmanTree(freq map[rune]int) (*Node, error) {
	if len(freq) == 0 {
		return nil, fmt.Errorf("frequency map is empty")
	}

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
		return heap.Pop(&pq).(*Node), nil // root of Huffman tree
	}
	return nil, fmt.Errorf("failed to build Huffman tree")
}

// buildHuffmanCodes builds Huffman codes (bit strings) for each character
func buildHuffmanCodes(root *Node) (map[rune]string, error) {
	if root == nil {
		return nil, fmt.Errorf("cannot build codes from nil root")
	}

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
	return codes, nil
}

// rebuildHuffmanTree reconstructs the Huffman tree from Huffman codes
func rebuildHuffmanTree(root *Node, codes map[rune]string) error {
	for char, code := range codes {
		insertNode(root, char, code)
	}
	return nil
}

// insertNode inserts a node into the Huffman tree based on the given code
func insertNode(root *Node, char rune, code string) {
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

// Zip compresses files using Huffman coding and returns a compressed file object.
func Zip(files []utils.File) (utils.File, error) {
	var buf bytes.Buffer

	// Write the number of files in the header
	err := binary.Write(&buf, binary.BigEndian, uint32(len(files)))
	if err != nil {
		return utils.File{}, err
	}

	// Create raw content buffer
	var rawContent bytes.Buffer
	err = createContentBuffer(files, &rawContent)
	if err != nil {
		return utils.File{}, err
	}

	// Compress rawContent using Huffman coding
	freq := make(map[rune]int)
	for _, b := range rawContent.Bytes() {
		freq[rune(b)]++
	}
	root, err := buildHuffmanTree(freq)
	if err != nil {
		return utils.File{}, err
	}
	codes, err := buildHuffmanCodes(root)
	if err != nil {
		return utils.File{}, err
	}
	compressedContent, err := compressData(rawContent.Bytes(), codes)
	if err != nil {
		return utils.File{}, err
	}

	// Write compressed content length to buffer
	err = binary.Write(&buf, binary.BigEndian, uint32(len(compressedContent)))
	if err != nil {
		return utils.File{}, err
	}
	// Write compressed content to buffer
	_, err = buf.Write(compressedContent)
	if err != nil {
		return utils.File{}, err
	}

	// Write Huffman codes length to buffer
	err = binary.Write(&buf, binary.BigEndian, uint32(len(codes)))
	if err != nil {
		return utils.File{}, err
	}
	// Write Huffman codes to buffer
	err = writeHuffmanCodesToBuffer(codes, &buf)
	if err != nil {
		return utils.File{}, err
	}

	return utils.File{
		Name:    "compressed.bin",
		Content: buf.Bytes(),
	}, nil
}

func writeHuffmanCodesToBuffer(codes map[rune]string, buf *bytes.Buffer) error {
	for char, code := range codes {
		err := binary.Write(buf, binary.BigEndian, char)
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.BigEndian, uint32(len(code)))
		if err != nil {
			return err
		}
		_, err = buf.WriteString(code)
		if err != nil {
			return err
		}
	}
	return nil
}

func createContentBuffer(files []utils.File, rawContent *bytes.Buffer) error {
	for _, file := range files {
		// Write filename length and filename
		filenameLen := uint32(len(file.Name))
		err := binary.Write(rawContent, binary.BigEndian, filenameLen)
		if err != nil {
			return err
		}
		_, err = rawContent.WriteString(file.Name)
		if err != nil {
			return err
		}

		// Write content length and content
		contentLen := uint32(len(file.Content))
		err = binary.Write(rawContent, binary.BigEndian, contentLen)
		if err != nil {
			return err
		}
		_, err = rawContent.Write(file.Content)
		if err != nil {
			return err
		}
	}
	return nil
}

// Unzip decompresses a compressed file using Huffman coding and returns individual files.
func Unzip(file utils.File) ([]utils.File, error) {
	var files []utils.File
	// Read file content
	buf := bytes.NewBuffer(file.Content)

	// Read number of files in header
	var numFiles uint32
	err := binary.Read(buf, binary.BigEndian, &numFiles)
	if err != nil {
		fmt.Printf("Error reading number of files: %v\n", err)
		return nil, err
	}

	// Read compressed content length
	var compressedContentLength uint32
	err = binary.Read(buf, binary.BigEndian, &compressedContentLength)
	if err != nil {
		return nil, err
	}
	compressedContent := make([]byte, compressedContentLength)
	_, err = buf.Read(compressedContent)
	if err != nil {
		return nil, err
	}

	// Read Huffman codes length
	var codesLength uint32
	err = binary.Read(buf, binary.BigEndian, &codesLength)
	if err != nil {
		return nil, err
	}

	codes := make(map[rune]string)
	// Read Huffman codes
	readHuffManCodes(&codes, buf, codesLength)

	// Rebuild Huffman tree using codes
	var root Node
	if len(codes) > 0 {
		err = rebuildHuffmanTree(&root, codes)
		if err != nil {
			return nil, err
		}
	}

	decompressedContent := decompressData(compressedContent, &root)

	// Parse decompressed content to extract files
	decompressedContentBuf := bytes.NewBuffer(decompressedContent)
	err = parseDecompressedContent(&files, &numFiles, decompressedContentBuf)

	if err != nil {
		return nil, err
	}

	return files, nil
}

func parseDecompressedContent(files *[]utils.File, numFiles *uint32, decompressedContentBuf *bytes.Buffer) error {
	for f := uint32(0); f < *numFiles; f++ {
		// Read filename length
		var nameLength uint32
		err := binary.Read(decompressedContentBuf, binary.BigEndian, &nameLength)
		if err != nil {
			return err
		}
		// Read filename
		name := make([]byte, nameLength)
		_, err = decompressedContentBuf.Read(name)
		if err != nil {
			return err
		}
		// Read content length
		var contentLength uint32
		err = binary.Read(decompressedContentBuf, binary.BigEndian, &contentLength)
		if err != nil {
			return err
		}
		// Read content
		content := make([]byte, contentLength)
		_, err = decompressedContentBuf.Read(content)
		if err != nil {
			return err
		}
		*files = append(*files, utils.File{
			Name:    string(name),
			Content: content,
		})
	}
	return nil
}

func readHuffManCodes(codes *map[rune]string, buf *bytes.Buffer, codesLength uint32) error {
	for i := uint32(0); i < codesLength; i++ {
		var char rune
		err := binary.Read(buf, binary.BigEndian, &char)
		if err != nil {
			return err
		}
		var codeLength uint32
		err = binary.Read(buf, binary.BigEndian, &codeLength)
		if err != nil {
			return err
		}
		code := make([]byte, codeLength)
		_, err = buf.Read(code)
		if err != nil {
			return err
		}
		(*codes)[char] = string(code)
	}
	return nil
}

// compressData compresses data using Huffman codes.
func compressData(data []byte, codes map[rune]string) ([]byte, error) {
	var buf bytes.Buffer
	var bitBuffer uint64
	var bitLength uint
	for _, b := range data {
		code := codes[rune(b)]
		err := compressBits(&code, &buf, &bitBuffer, &bitLength)
		if err != nil {
			return nil, err
		}
	}
	if bitLength > 0 {
		bitBuffer <<= (64 - bitLength)
		err := binary.Write(&buf, binary.BigEndian, bitBuffer)
		if err != nil {
			return nil, fmt.Errorf("failed to write bit buffer: %v", err)
		}
	}
	return buf.Bytes(), nil
}

func compressBits(code *string, buf *bytes.Buffer, bitBuffer *uint64, bitLength *uint) error {
	for _, bit := range *code {
		*bitBuffer <<= 1
		*bitLength++
		if bit == '1' {
			*bitBuffer |= 1
		}
		if *bitLength == 64 {
			err := binary.Write(buf, binary.BigEndian, bitBuffer)
			if err != nil {
				return fmt.Errorf("failed to write bit buffer: %v", err)
			}
			*bitBuffer = 0
			*bitLength = 0
		}
	}
	return nil
}

// decompressData decompresses data using Huffman codes.
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
