package compressor

import (
    "bytes"
    "container/heap"
    "encoding/binary"
    "io"
)

// Node represents a node in the Huffman tree.
type Node struct {
    char  rune
    freq  int
    left  *Node
    right *Node
}

// PriorityQueue implements heap.Interface and holds Nodes.
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
    *pq = old[0 : n-1]
    return item
}

// buildHuffmanTree builds the Huffman tree from the frequency table.
func buildHuffmanTree(freq map[rune]int) *Node {
    pq := make(PriorityQueue, len(freq))
    i := 0
    for char, f := range freq {
        pq[i] = &Node{char: char, freq: f}
        i++
    }
    heap.Init(&pq)

    for pq.Len() > 1 {
        left := heap.Pop(&pq).(*Node)
        right := heap.Pop(&pq).(*Node)
        parent := &Node{
            freq:  left.freq + right.freq,
            left:  left,
            right: right,
        }
        heap.Push(&pq, parent)
    }

    return heap.Pop(&pq).(*Node)
}

// buildHuffmanCodes builds the Huffman codes from the Huffman tree.
func buildHuffmanCodes(root *Node) map[rune]string {
    codes := make(map[rune]string)
    var buildCodes func(*Node, string)
    buildCodes = func(node *Node, code string) {
        if node == nil {
            return
        }
        if node.char != 0 {
            codes[node.char] = code
        }
        buildCodes(node.left, code+"0")
        buildCodes(node.right, code+"1")
    }
    buildCodes(root, "")
    return codes
}

// Compress compresses the input data using Huffman coding.
func Compress(data []byte) ([]byte, map[rune]string) {
    freq := make(map[rune]int)
    for _, b := range data {
        freq[rune(b)]++
    }

    root := buildHuffmanTree(freq)
    codes := buildHuffmanCodes(root)

    // Encode data using Huffman codes
    var buffer bytes.Buffer
    bitBuffer := byte(0)
    bitCount := 0

    for _, b := range data {
        code := codes[rune(b)]
        for _, bit := range code {
            bitBuffer <<= 1
            if bit == '1' {
                bitBuffer |= 1
            }
            bitCount++
            if bitCount == 8 {
                buffer.WriteByte(bitBuffer)
                bitBuffer = 0
                bitCount = 0
            }
        }
    }

    // Write remaining bits
    if bitCount > 0 {
        bitBuffer <<= (8 - bitCount)
        buffer.WriteByte(bitBuffer)
    }

    // Write the original uncompressed data size as the first 4 bytes
    sizeBuffer := make([]byte, 4)
    binary.LittleEndian.PutUint32(sizeBuffer, uint32(len(data)))
    buffer.Write(sizeBuffer)

    return buffer.Bytes(), codes
}

// Decompress decompresses the input data using Huffman coding.
func Decompress(data []byte, codes map[rune]string) []byte {
    reader := bytes.NewReader(data)
    uncompressedSizeBytes := make([]byte, 4)
    if _, err := io.ReadFull(reader, uncompressedSizeBytes); err != nil {
        return nil
    }
    uncompressedSize := int(binary.LittleEndian.Uint32(uncompressedSizeBytes))

    var decodedData []byte
    currentCode := ""
    for {
        bit, err := reader.ReadByte()
        if err == io.EOF {
            break
        }
        for i := 7; i >= 0; i-- {
            if (bit>>i)&1 == 1 {
                currentCode += "1"
            } else {
                currentCode += "0"
            }
            for char, code := range codes {
                if currentCode == code {
                    decodedData = append(decodedData, byte(char))
                    currentCode = ""
                    break
                }
            }
        }
    }

    return decodedData[:uncompressedSize]
}
