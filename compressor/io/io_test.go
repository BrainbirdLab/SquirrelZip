package io

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestArray(t *testing.T) {
	arr := make([]byte, 2)
	arr[0] = 1
	arr[1] = 2
	fmt.Println(arr)
	//add more elements to the array
	arr = append(arr, 3)
	fmt.Println(arr)
}

func WriteMap(data map[string]int, key string, value int) {
	data[key] = value
}

func TestWriteMap(t *testing.T) {
	data := make(map[string]int)
	WriteMap(data, "key", 10)
	fmt.Println(data)
	WriteMap(data, "key", 20)
	fmt.Println(data)
}

func TestRecursiveFolderRead(t *testing.T) {
	//get all filenames from a dir. include subdirectories
	target := "test_files"

	files := make([]string, 0)

	//use walk to get all files in the directory
	err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk directory: %v", err)
	}

	fmt.Printf("Files found: %v\n", files)
}

func TestRecursiveFolderWrite(t *testing.T) {
	path := "recursive_test/test1/test2/test3/file.txt"
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	file.Close()
	//delete the recursive_test directory
	defer os.RemoveAll("recursive_test")
}