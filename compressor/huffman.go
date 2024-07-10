package compressor

import (
	"bytes"
	"encoding/binary"
	"file-compressor/utils"
)

// Node represents a node in the Huffman tree
type Node struct {
	char  rune           // Character stored in the node
	freq  int            // Frequency of the character
	left  *Node          // Left child node
	right *Node          // Right child node
}

// PriorityQueue implements a priority queue for Nodes
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
func (pq *PriorityQueue) Push(x *Node) {
	*pq = append(*pq, x)
}

// Pop removes and returns the highest priority item (Node) from the priority queue
func (pq *PriorityQueue) Pop() *Node {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// buildHuffmanTree builds the Huffman tree from character frequencies
func buildHuffmanTree(freq map[rune]int) *Node {
	pq := make(PriorityQueue, len(freq))
	i := 0
	for char, f := range freq {
		pq[i] = &Node{char: char, freq: f}
		i++
	}
	buildMinHeap(pq)

	for len(pq) > 1 {
		left := pq.Pop()
		right := pq.Pop()

		internal := &Node{
			char:  '\x00', // internal node character
			freq:  left.freq + right.freq,
			left:  left,
			right: right,
		}
		pq.Push(internal)
	}

	if len(pq) > 0 {
		return pq.Pop() // root of Huffman tree
	}
	return nil
}

// buildMinHeap builds a min-heap for the priority queue
func buildMinHeap(pq PriorityQueue) {
	n := len(pq)
	for i := n/2 - 1; i >= 0; i-- {
		heapify(pq, n, i)
	}
}

// heapify maintains the heap property of the priority queue
func heapify(pq PriorityQueue, n, i int) {
	smallest := i
	left := 2*i + 1
	right := 2*i + 2

	if left < n && pq[left].freq < pq[smallest].freq {
		smallest = left
	}

	if right < n && pq[right].freq < pq[smallest].freq {
		smallest = right
	}

	if smallest != i {
		pq.Swap(i, smallest)
		heapify(pq, n, smallest)
	}
}

// buildHuffmanCodes builds Huffman codes (bit strings) for each character
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

// rebuildHuffmanTree reconstructs the Huffman tree from Huffman codes
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


// Zip compresses files using Huffman coding and returns a compressed file object.
func Zip(files []utils.File) utils.File {
	var buf bytes.Buffer

	// Write the number of files in the header
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
		Name:    "compressed",
		Content: buf.Bytes(),
	}
}

// Unzip decompresses a compressed file using Huffman coding and returns individual files.
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


// compressData compresses data using Huffman codes.
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
