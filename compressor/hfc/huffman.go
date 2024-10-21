package hfc

import (
	"container/heap"
	"errors"
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

	// put all nodes in the priority queue
	for len(pq) > 1 {
		left := heap.Pop(&pq).(*Node)
		right := heap.Pop(&pq).(*Node)
		internal := &Node{freq: left.freq + right.freq, left: left, right: right}
		heap.Push(&pq, internal)
	}

	return heap.Pop(&pq).(*Node), nil
}


// GetHuffmanCodes generates Huffman codes for the given frequency map of runes.
// It builds a Huffman tree based on the frequencies and then traverses the tree
// to generate the corresponding Huffman codes.
//
// Parameters:
//   - freq: A pointer to a map where keys are runes and values are their frequencies.
//
// Returns:
//   - A map where keys are runes and values are their corresponding Huffman codes.
//   - An error if there is an issue building the Huffman tree.
func GetHuffmanCodes(freq *map[rune]int) (map[rune]string, error) {

	codes := make(map[rune]string)

	node, err := buildHuffmanTree(freq)
	if err != nil {
		return nil, err
	}

	huffmanBuilder(node, "", &codes, freq)

	return codes, nil
}


// huffmanBuilder recursively builds the Huffman codes for each character in the input.
// It traverses the Huffman tree and assigns binary codes to characters based on their position in the tree.
//
// Parameters:
//   - node: A pointer to the current node in the Huffman tree.
//   - prefix: The current binary prefix string representing the path taken to reach the node.
//   - codes: A pointer to a map that stores the Huffman codes for each character.
//   - frequency: A pointer to a map that stores the frequency of each character (not used in this function).
//
// If the current node is a leaf node (both left and right children are nil), the function assigns the current
// prefix to the character stored in the node. Otherwise, it recursively traverses the left and right children,
// appending '0' to the prefix for the left child and '1' for the right child.
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

// rebuildHuffmanTree reconstructs a Huffman tree from a given map of runes to their corresponding binary codes.
// Each rune in the map represents a character, and the associated string represents the binary code for that character.
// The function returns the root node of the reconstructed Huffman tree.
//
// Parameters:
//   codes (map[rune]string): A map where keys are runes (characters) and values are strings representing the binary codes.
//
// Returns:
//   *Node: The root node of the reconstructed Huffman tree.
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