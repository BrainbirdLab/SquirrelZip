package compressor

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"path/filepath"

	"file-compressor/utils"
)

// Compress compresses the specified files or folders into a single compressed file.
func Compress(filenameStrs []string, outputDir *string, password *string) error {
	// Check if there are files or folders to compress
	if len(filenameStrs) == 0 {
		return errors.New("no files or folders to compress")
	}

	// Set default output directory if not provided
	if *outputDir == "" {
		*outputDir = path.Dir(filenameStrs[0])
	}

	// Prepare to store files' content
	var files []utils.File

	// Use a wait group to synchronize goroutines
	var wg sync.WaitGroup
	var errMutex sync.Mutex // Mutex to handle errors safely

	// Channel to receive errors from goroutines
	errChan := make(chan error, len(filenameStrs))

	// Process each input file or folder
	for _, filename := range filenameStrs {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()

			// Check if the file or folder exists
			info, err := os.Stat(filename)
			if os.IsNotExist(err) {
				errChan <- fmt.Errorf("file or folder does not exist: %s", filename)
				return
			}

			// Handle directory recursively
			if info.IsDir() {
				err := compressFolderRecursive(filename, &files)
				if err != nil {
					errChan <- err
				}
			} else {
				// Read file content
				content, err := os.ReadFile(filename)
				if err != nil {
					errChan <- fmt.Errorf("failed to read file %s: %w", filename, err)
					return
				}

				// Store file information (name and content)
				files = append(files, utils.File{
					Name:    path.Base(filename),
					Content: content,
				})
			}
		}(filename)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check for errors from goroutines
	for err := range errChan {
		if err != nil {
			errMutex.Lock()
			defer errMutex.Unlock()
			return err // Return the first error encountered
		}
	}

	// Compress files using Huffman coding
	compressedFile := Zip(files)

	// Encrypt compressed content if password is provided
	if password != nil && *password != "" {
		var err error
		compressedFile.Content, err = utils.Encrypt(compressedFile.Content, *password)
		if err != nil {
			return fmt.Errorf("encryption error: %w", err)
		}
	}

	// Write compressed file to the output directory
	err := os.WriteFile(*outputDir+"/"+compressedFile.Name, compressedFile.Content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write compressed file: %w", err)
	}

	return nil
}

// Function to recursively compress a folder and its contents
func compressFolderRecursive(folderPath string, files *[]utils.File) error {
	// Traverse the folder contents
	err := filepath.Walk(folderPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Read file content
			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, err)
			}

			filename, err := filepath.Rel(folderPath, filePath)
			if err != nil {
				return fmt.Errorf("failed to get relative path for file %s: %w", filePath, err)
			}

			// Store file information (relative path and content)
			*files = append(*files, utils.File{
				Name:    filepath.ToSlash(filename), // Store relative path
				Content: content,
			})
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error compressing folder %s: %w", folderPath, err)
	}

	return nil
}

// Decompress decompresses the specified compressed file into individual files or folders.
func Decompress(filenameStrs []string, outputDir *string, password *string) error {
	// Check if there are files to decompress
	if len(filenameStrs) == 0 {
		return errors.New("no files to decompress")
	}

	// Set default output directory if not provided
	if *outputDir == "" {
		*outputDir = path.Dir(filenameStrs[0])
	}

	// Read compressed file content
	compressedContent, err := os.ReadFile(filenameStrs[0])
	if err != nil {
		return fmt.Errorf("failed to read compressed file %s: %w", filenameStrs[0], err)
	}

	// Decrypt compressed content if password is provided
	if password != nil && *password != "" {
		compressedContent, err = utils.Decrypt(compressedContent, *password)
		if err != nil {
			return fmt.Errorf("decryption error: %w", err)
		}
	}

	// Decompress file using Huffman coding
	files := Unzip(utils.File{
		Name:    path.Base(filenameStrs[0]),
		Content: compressedContent,
	})

	// Use a wait group to synchronize goroutines
	var wg sync.WaitGroup
	var errMutex sync.Mutex // Mutex to handle errors safely

	// Channel to receive errors from goroutines
	errChan := make(chan error, len(files))

	// Process each decompressed file
	for _, file := range files {
		wg.Add(1)
		go func(file utils.File) {
			defer wg.Done()

			// Create directories if they don't exist
			outputPath := filepath.Join(*outputDir, filepath.Dir(file.Name))
			err := os.MkdirAll(outputPath, os.ModePerm)
			if err != nil {
				errChan <- fmt.Errorf("failed to create directory %s: %w", outputPath, err)
				return
			}

			// Write decompressed file content
			err = os.WriteFile(filepath.Join(*outputDir, file.Name), file.Content, 0644)
			if err != nil {
				errChan <- fmt.Errorf("failed to write decompressed file %s: %w", file.Name, err)
				return
			}
			utils.ColorPrint(utils.GREEN, fmt.Sprintf("Decompressed file: %s\n", file.Name))
		}(file)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	// Check for errors from goroutines
	for err := range errChan {
		if err != nil {
			errMutex.Lock()
			defer errMutex.Unlock()
			return err // Return the first error encountered
		}
	}

	return nil
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
		Name:    "compressed.bin",
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
