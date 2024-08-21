package hfc

import (
	"container/heap"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
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

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].freq < pq[j].freq }
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }

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

// buildHuffmanCodes builds Huffman codes for each character.
func buildHuffmanCodes(node *Node, prefix string, codes map[rune]string) {
	if node == nil {
		return
	}
	if node.left == nil && node.right == nil {
		codes[node.char] = prefix
		return
	}
	buildHuffmanCodes(node.left, prefix+"0", codes)
	buildHuffmanCodes(node.right, prefix+"1", codes)
}

// compressData compresses data using Huffman codes and writes the compressed contents to the output.
func compressData(input io.Reader, output io.Writer, codes *map[rune]string) error {
	var bitBuffer byte
	var bitCount int

	buf := make([]byte, 512)
	for {
		n, err := input.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break // EOF reached
		}

		if err := processCompressionBuffer(buf[:n], codes, &bitBuffer, &bitCount, output); err != nil {
			return err
		}
	}

	// Write any remaining bits in the buffer
	if bitCount > 0 {
		if _, err := output.Write([]byte{bitBuffer}); err != nil {
			return fmt.Errorf("error writing final byte: %w", err)
		}
	}

	return nil
}

func processCompressionBuffer(buf []byte, codes *map[rune]string, bitBuffer *byte, bitCount *int, output io.Writer) error {
	for _, b := range buf {
		code, ok := (*codes)[rune(b)]
		if !ok {
			return fmt.Errorf("character not found in Huffman codes: %c", b)
		}
		for _, bit := range code {
			if bit == '1' {
				*bitBuffer |= 1 << (7 - *bitCount)
			}
			*bitCount++

			// When bitCount reaches 8, the buffer is full
			if *bitCount == 8 {
				if _, err := output.Write([]byte{*bitBuffer}); err != nil {
					return fmt.Errorf("error writing byte: %w", err)
				}
				*bitBuffer = 0
				*bitCount = 0
			}
		}
	}

	return nil
}



// decompressData decompresses data using the Huffman tree in a streaming manner.
func decompressData(input io.Reader, output io.Writer, root *Node) error {
	buf := make([]byte, 256)
	node := root

	for {
		n, err := input.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading input: %w", err)
		}
		if n == 0 {
			break // EOF reached
		}

		if err := processDecompressionBuffer(buf[:n], node, output, root); err != nil {
			return err
		}
	}

	return nil
}

func processDecompressionBuffer(buf []byte, node *Node, output io.Writer, root *Node) error {
	for _, b := range buf {
		for i := 7; i >= 0; i-- {
			bit := (b >> i) & 1
			if bit == 0 {
				node = node.left
			} else {
				node = node.right
			}

			if node.left == nil && node.right == nil {
				if _, err := output.Write([]byte{byte(node.char)}); err != nil {
					return fmt.Errorf("error writing character: %w", err)
				}
				node = root
			}
		}
	}

	return nil
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
		if err := binary.Write(output, binary.LittleEndian, char); err != nil {
			return fmt.Errorf("error writing character: %w", err)
		}
		if err := binary.Write(output, binary.LittleEndian, int32(len(code))); err != nil {
			return fmt.Errorf("error writing code length: %w", err)
		}
		if _, err := output.Write([]byte(code)); err != nil {
			return fmt.Errorf("error writing code: %w", err)
		}
	}

	return nil
}



func readHuffmanCodes(input io.Reader, codes *map[rune]string) error {
	var numCodes int32
	if err := binary.Read(input, binary.LittleEndian, &numCodes); err != nil {
		return fmt.Errorf("error reading number of codes: %w", err)
	}

	for i := 0; i < int(numCodes); i++ {
		var char rune
		if err := binary.Read(input, binary.LittleEndian, &char); err != nil {
			return fmt.Errorf("error reading character: %w", err)
		}

		var codeLen int32
		if err := binary.Read(input, binary.LittleEndian, &codeLen); err != nil {
			return fmt.Errorf("error reading code length: %w", err)
		}

		code := make([]byte, codeLen)
		if _, err := input.Read(code); err != nil {
			return fmt.Errorf("error reading code: %w", err)
		}

		(*codes)[char] = string(code)
	}

	return nil
}




// Zip compresses data using Huffman coding and writes the compressed data to the output stream.
func Zip(input io.Reader, output io.Writer) error {
	// Step 1: Get frequency map
	freq := make(map[rune]int)
	if err := getFrequencyMap(input, &freq); err != nil {
		return fmt.Errorf("error generating frequency map: %w", err)
	}
	// Reset input reader to start reading data again
	if seeker, ok := input.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("error resetting input reader: %w", err)
		}
	} else {
		return fmt.Errorf("input reader does not support seeking")
	}

	// Step 2: Build Huffman tree and codes
	root, err := buildHuffmanTree(&freq)
	if err != nil {
		return fmt.Errorf("error building Huffman tree: %w", err)
	}

	codes := make(map[rune]string)
	buildHuffmanCodes(root, "", codes)

	// Step 3: Write frequency map and Huffman codes to the output
	if err := writeHuffmanCodes(codes, output); err != nil {
		return fmt.Errorf("error writing Huffman codes: %w", err)
	}

	// Step 4: Compress and write the data
	if err := compressData(input, output, &codes); err != nil {
		return fmt.Errorf("error compressing data: %w", err)
	}

	fmt.Printf("Data compressed successfully\n")

	return nil
}


// Unzip decompresses data using Huffman coding and writes the decompressed data to the output stream.
func Unzip(input io.Reader, output io.Writer) error {
	codes := make(map[rune]string)
	if err := readHuffmanCodes(input, &codes); err != nil {
		return fmt.Errorf("error reading Huffman codes: %w", err)
	}

	root := &Node{}
	rebuildHuffmanTree(&codes, &root)

	if err := decompressData(input, output, root); err != nil {
		return fmt.Errorf("error decompressing data: %w", err)
	}

	return nil
}

func rebuildHuffmanTree(codes *map[rune]string, root **Node) {
	for char, code := range *codes {
		node := *root
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
}
