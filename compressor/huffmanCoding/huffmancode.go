package huffmanCoding

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// Node represents a node in the Huffman tree
type Node struct {
	Char  rune
	Freq  int
	Left  *Node
	Right *Node
}

// PriorityQueue implements a priority queue for nodes
type PriorityQueue []*Node

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].Freq < pq[j].Freq }
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }

func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*Node))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}

// BuildHuffmanTree constructs a Huffman tree from a frequency map
func BuildHuffmanTree(freqMap map[rune]int) *Node {
	pq := &PriorityQueue{}
	heap.Init(pq)

	for char, freq := range freqMap {
		heap.Push(pq, &Node{Char: char, Freq: freq})
	}

	for pq.Len() > 1 {
		left := heap.Pop(pq).(*Node)
		right := heap.Pop(pq).(*Node)
		merged := &Node{
			Freq:  left.Freq + right.Freq,
			Left:  left,
			Right: right,
		}
		heap.Push(pq, merged)
	}

	if pq.Len() == 0 {
		return nil
	}
	return heap.Pop(pq).(*Node)
}

// BuildHuffmanCodes generates Huffman codes from the Huffman tree
func BuildHuffmanCodes(node *Node, prefix string, codes map[rune]string) {
	if node == nil {
		return
	}
	if node.Left == nil && node.Right == nil {
		codes[node.Char] = prefix
		return
	}
	BuildHuffmanCodes(node.Left, prefix+"0", codes)
	BuildHuffmanCodes(node.Right, prefix+"1", codes)
}

// Encode encodes data using the Huffman codes
func Encode(data []byte, codes map[rune]string) ([]byte, error) {
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


// Decode decodes Huffman encoded data
func Decode(encodedData []byte, root *Node) ([]byte, error) {
	var decodedData bytes.Buffer
	node := root

	bits := binaryStringFromByteSlice(encodedData)
	for _, bit := range bits {
		if bit == '0' {
			node = node.Left
		} else {
			node = node.Right
		}
		if node.Left == nil && node.Right == nil {
			decodedData.WriteRune(node.Char)
			node = root
		}
	}

	return decodedData.Bytes(), nil
}




// Zip compresses data from an io.Reader and writes it to an io.Writer
func Zip(input io.Reader, output io.Writer) error {
	data, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	freqMap := make(map[rune]int)
	for _, b := range data {
		freqMap[rune(b)]++
	}

	root := BuildHuffmanTree(freqMap)
	if root == nil {
		return fmt.Errorf("failed to build Huffman tree")
	}

	codes := make(map[rune]string)
	BuildHuffmanCodes(root, "", codes)

	encodedData, err := Encode(data, codes)
	if err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	fmt.Printf("Encoded data len %d\n", len(encodedData))
	fmt.Printf("Freq map len %d\n", len(freqMap))
	fmt.Printf("codes map len %d\n", len(codes))

	
	// Write frequency map size
	err = binary.Write(output, binary.BigEndian, int32(len(freqMap)))
	if err != nil {
		return fmt.Errorf("failed to write frequency map size: %w", err)
	}

	// Write frequency map
	for char, freq := range freqMap {
		err = binary.Write(output, binary.BigEndian, int32(char))
		if err != nil {
			return fmt.Errorf("failed to write character: %w", err)
		}
		err = binary.Write(output, binary.BigEndian, int32(freq))
		if err != nil {
			return fmt.Errorf("failed to write frequency: %w", err)
		}
	}

	// Write encoded data
	_, err = output.Write(encodedData)
	if err != nil {
		return fmt.Errorf("failed to write compressed data: %w", err)
	}

	return nil
}

func Unzip(input io.Reader, output io.Writer) error {
	var freqMap = make(map[rune]int)
	var size int32

	// Read frequency map size
	err := binary.Read(input, binary.BigEndian, &size)
	if err != nil {
		return fmt.Errorf("failed to read frequency map size: %w", err)
	}

	// Read frequency map
	for i := int32(0); i < size; i++ {
		var char int32
		var freq int32
		err = binary.Read(input, binary.BigEndian, &char)
		if err != nil {
			return fmt.Errorf("failed to read character: %w", err)
		}
		err = binary.Read(input, binary.BigEndian, &freq)
		if err != nil {
			return fmt.Errorf("failed to read frequency: %w", err)
		}
		freqMap[rune(char)] = int(freq)
	}

	root := BuildHuffmanTree(freqMap)
	if root == nil {
		return fmt.Errorf("failed to build Huffman tree")
	}

	encodedData, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read compressed data: %w", err)
	}

	decodedData, err := Decode(encodedData, root)
	if err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}

	_, err = output.Write(decodedData)
	if err != nil {
		return fmt.Errorf("failed to write decompressed data: %w", err)
	}

	return nil
}


// Helper function to convert binary string to byte slice
func byteSliceFromBinaryString(s string) []byte {
	var result []byte
	// Ensure the string length is a multiple of 8
	if len(s)%8 != 0 {
		s = s + strings.Repeat("0", 8-len(s)%8)
	}
	for i := 0; i < len(s); i += 8 {
		var byteValue byte
		for j := 0; j < 8; j++ {
			if s[i+j] == '1' {
				byteValue |= 1 << (7 - j)
			}
		}
		result = append(result, byteValue)
	}
	return result
}



func binaryStringFromByteSlice(data []byte) string {
	var buffer bytes.Buffer
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			if (b>>i)&1 == 1 {
				buffer.WriteByte('1')
			} else {
				buffer.WriteByte('0')
			}
		}
	}
	return buffer.String()
}

