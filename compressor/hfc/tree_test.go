package hfc

import (
	"bytes"
	"fmt"
	"testing"
)


func TestTree(t *testing.T) {
	testData := []byte("lorem ipsum dolor sit amet, consectetur adipiscing elit. sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	reader := bytes.NewBuffer(testData)
	//make frequency map
	freq := make(map[rune]int)
	getFrequencyMap(reader, &freq)

	//make tree
	root, err := buildHuffmanTree(&freq)
	if err != nil {
		t.Fatalf("failed to build huffman tree: %v", err)
	}

	//make codes
	codes := make(map[rune]string)
	huffmanBuilder(root, "", &codes, &freq)

	//rebuild tree from codes
	root2 := rebuildHuffmanTree(codes)

	//compare the trees
	err = compareHuffmanTrees(t, root, root2)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Equal trees")
}

func compareHuffmanTrees(t *testing.T, root1, root2 *Node) error {
	if root1 == nil && root2 == nil {
		return nil
	}

	if root1 == nil || root2 == nil {
		return fmt.Errorf("trees are not equal. one of the nodes is nil")
	}

	err := compareHuffmanTrees(t, root1.left, root2.left)
	if err != nil {
		return err
	}

	err = compareHuffmanTrees(t, root1.right, root2.right)
	if err != nil {
		return err
	}

	return nil
}
