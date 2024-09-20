package hfc

import (
	"bytes"
	"io"
	"os"
	"testing"

	ec "file-compressor/errorConstants"
)

func PrintError(t *testing.T, msg string, err error) {
	if err != nil {
		t.Fatalf(msg, err)
	}
}

func TestSimpleString(t *testing.T) {
	data := []byte("aaaabbbbccccdddd")
	reader := bytes.NewBuffer(data)
	freq := make(map[rune]int)
	if err := getFrequencyMap(reader, &freq); err != nil {
		PrintError(t, ec.FAILED_GET_FREQ_MAP, err)
	}

	codes, err := GetHuffmanCodes(&freq)
	if err != nil {
		PrintError(t, ec.FAILED_BUILD_HUFFMAN_CODES, err)
	}

	file, err := os.Create("simple_huffman_codes.txt")
	if err != nil {
		PrintError(t, ec.FILE_CREATE_ERROR, err)
	}
	
	if err := WriteHuffmanCodes(file, codes); err != nil {
		PrintError(t, ec.FAILED_WRITE_HUFFMAN_CODES, err)
	}
	
	file.Seek(0, io.SeekStart) // Move file pointer back to the beginning
	
	codes2, err := ReadHuffmanCodes(file)
	if err != nil {
		PrintError(t, ec.FAILED_READ_HUFFMAN_CODES, err)
	}
	
	compareHuffmanCodes(t, codes, codes2)
	
	file.Close()
	//remove the file
	if err := os.Remove("simple_huffman_codes.txt"); err != nil {
		PrintError(t, ec.FILE_REMOVE_ERROR, err)
	}
}


func TestUniqueCharacters(t *testing.T) {
	data := []byte("abcdefg")
	reader := bytes.NewBuffer(data)
	freq := make(map[rune]int)
	if err := getFrequencyMap(reader, &freq); err != nil {
		PrintError(t, ec.FAILED_GET_FREQ_MAP, err)
	}

	codes, err := GetHuffmanCodes(&freq)
	if err != nil {
		PrintError(t, ec.FAILED_BUILD_HUFFMAN_CODES, err)
	}

	file, err := os.Create("unique_huffman_codes.txt")
	if err != nil {
		PrintError(t, ec.FILE_CREATE_ERROR, err)
	}
	
	if err := WriteHuffmanCodes(file, codes); err != nil {
		PrintError(t, ec.FAILED_WRITE_HUFFMAN_CODES, err)
	}

	file.Seek(0, io.SeekStart)
	
	codes2, err := ReadHuffmanCodes(file)
	if err != nil {
		PrintError(t, ec.FAILED_READ_HUFFMAN_CODES, err)
	}
	
	compareHuffmanCodes(t, codes, codes2)
	
	file.Close()
	//remove the file
	if err := os.Remove("unique_huffman_codes.txt"); err != nil {
		PrintError(t, ec.FILE_REMOVE_ERROR, err)
	}
}

func TestSingleCharacterRepeated(t *testing.T) {
	data := []byte("aaaaaaa")
	reader := bytes.NewBuffer(data)
	freq := make(map[rune]int)
	if err := getFrequencyMap(reader, &freq); err != nil {
		PrintError(t, ec.FAILED_GET_FREQ_MAP, err)
	}

	codes, err := GetHuffmanCodes(&freq)
	if err != nil {
		PrintError(t, ec.FAILED_BUILD_HUFFMAN_CODES, err)
	}

	file, err := os.Create("single_char_huffman_codes.txt")
	if err != nil {
		PrintError(t, ec.FILE_CREATE_ERROR, err)
	}
	
	if err := WriteHuffmanCodes(file, codes); err != nil {
		PrintError(t, ec.FAILED_WRITE_HUFFMAN_CODES, err)
	}

	file.Seek(0, io.SeekStart)
	
	codes2, err := ReadHuffmanCodes(file)
	if err != nil {
		PrintError(t, ec.FAILED_READ_HUFFMAN_CODES, err)
	}
	
	compareHuffmanCodes(t, codes, codes2)
	
	file.Close()
	//remove the file
	if err := os.Remove("single_char_huffman_codes.txt"); err != nil {
		PrintError(t, ec.FILE_REMOVE_ERROR, err)
	}
}

func TestLongText(t *testing.T) {
	data := []byte("this is a longer text to test the huffman encoding system with a more extensive input")
	reader := bytes.NewBuffer(data)
	freq := make(map[rune]int)
	if err := getFrequencyMap(reader, &freq); err != nil {
		PrintError(t, ec.FAILED_GET_FREQ_MAP, err)
	}

	codes, err := GetHuffmanCodes(&freq)
	if err != nil {
		PrintError(t, ec.FAILED_BUILD_HUFFMAN_CODES, err)
	}

	file, err := os.Create("long_text_huffman_codes.txt")
	if err != nil {
		PrintError(t, ec.FILE_CREATE_ERROR, err)
	}

	
	if err := WriteHuffmanCodes(file, codes); err != nil {
		PrintError(t, ec.FAILED_WRITE_HUFFMAN_CODES, err)
	}

	file.Seek(0, io.SeekStart)

	codes2, err := ReadHuffmanCodes(file)
	if err != nil {
		PrintError(t, ec.FAILED_READ_HUFFMAN_CODES, err)
	}
	
	compareHuffmanCodes(t, codes, codes2)
	
	file.Close()
	//remove the file
	if err := os.Remove("long_text_huffman_codes.txt"); err != nil {
		PrintError(t, ec.FILE_REMOVE_ERROR, err)
	}
}

func TestSpecialCharacters(t *testing.T) {
	data := []byte("hello, world!😠🤲🙈😁😒🥺😀\nThis is a test!")
	reader := bytes.NewBuffer(data)
	freq := make(map[rune]int)
	if err := getFrequencyMap(reader, &freq); err != nil {
		PrintError(t, ec.FAILED_GET_FREQ_MAP, err)
	}

	codes, err := GetHuffmanCodes(&freq)
	if err != nil {
		PrintError(t, ec.FAILED_BUILD_HUFFMAN_CODES, err)
	}

	file, err := os.Create("special_char_huffman_codes.txt")
	if err != nil {
		PrintError(t, ec.FILE_CREATE_ERROR, err)
	}
	
	if err := WriteHuffmanCodes(file, codes); err != nil {
		PrintError(t, ec.FAILED_WRITE_HUFFMAN_CODES, err)
	}

	file.Seek(0, io.SeekStart)
	
	codes2, err := ReadHuffmanCodes(file)
	if err != nil {
		PrintError(t, ec.FAILED_READ_HUFFMAN_CODES, err)
	}

	compareHuffmanCodes(t, codes, codes2)
	
	file.Close()
	//remove the file
	if err := os.Remove("special_char_huffman_codes.txt"); err != nil {
		PrintError(t, ec.FILE_REMOVE_ERROR, err)
	}
}


func compareHuffmanCodes(t *testing.T, codes1, codes2 map[rune]string) {
	if len(codes1) != len(codes2) {
		t.Fatalf("code lengths do not match: %d vs %d", len(codes1), len(codes2))
	}

	for k, v1 := range codes1 {
		v2, ok := codes2[k]
		if !ok {
			t.Fatalf("missing rune %v in codes2", k)
		}
		if v1 != v2 {
			t.Fatalf("codes for rune %v do not match: %s vs %s", k, v1, v2)
		}
	}
}