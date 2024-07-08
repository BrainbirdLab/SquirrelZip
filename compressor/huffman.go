package compressor

import (
	"container/heap"
)

// Node represents a node in the Huffman tree
type Node struct {
	char  rune    // Character stored in the node
	freq  int     // Frequency of the character
	left  *Node   // Left child node
	right *Node   // Right child node
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
