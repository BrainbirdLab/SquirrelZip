package compressor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"file-compressor/encryption"
	"file-compressor/utils"
	ar "file-compressor/compressor/arithmetic"
	hc "file-compressor/compressor/huffmanCoding"
)

func WriteCompressionAlgorithm(algo string, file *utils.File) {

	data := file.Content

	algoNameLen := len(algo)

	rawData := make([]byte, algoNameLen+1+len(data))

	rawData[0] = byte(algoNameLen)

	copy(rawData[1:], []byte(algo))

	copy(rawData[1+algoNameLen:], data)

	file.Content = rawData
}

func CheckCompressionAlgorithm(algo string) error {
	switch utils.Algorithm(algo) {
	case utils.HUFFMAN, utils.ARITHMETIC:
		return nil
	default:
		return fmt.Errorf("unsupported compression algorithm: %v", algo)
	}
}

func ReadCompressionAlgorithm(file *utils.File) (utils.Algorithm, error) {
	
	if len(file.Content) == 0 {
		return utils.UNSUPPORTED, fmt.Errorf("empty file content")
	}

	// Read algorithm name length
	algoNameLen := int(file.Content[0])

	// Read algorithm name
	algoName := string(file.Content[1 : 1+algoNameLen])

	// Read compressed content
	compressedContent := file.Content[1+algoNameLen:]

	// Update file content
	file.Content = compressedContent

	return utils.Algorithm(algoName), nil
}

// Compress compresses the specified files or folders into a single compressed file.
func Compress(filenameStrs []string, outputDir string, password string, algorithm string) error {
	// Set default output directory if not provided
	if outputDir == "" {
		outputDir = filepath.Dir(filenameStrs[0])
	}

	// Prepare to store files' content
	var files []utils.File
	var originalSize int64
	var compressedSize int64

	// Read all files concurrently
	errs := ReadAllFilesConcurrently(filenameStrs, &files, &originalSize)

	// Check for errors from goroutines
	if len(errs) > 0 {
		for _, err := range errs {
			utils.ColorPrint(utils.RED, fmt.Sprintf("Error: %v\n", err))
		}
		os.Exit(1)
	}


	// Check if the compression algorithm is supported
	err := CheckCompressionAlgorithm(algorithm)
	if err != nil {
		return err
	}

	var compressedFile utils.File

	switch utils.Algorithm(algorithm) {
	case utils.HUFFMAN:
		// Compress files using Huffman coding
		compressedFile, err = hc.Zip(files)
	case utils.ARITHMETIC:
		// Compress files using Arithmetic coding
		compressedFile, err = ar.Zip(files)
	}

	if err != nil {
		return err
	}

	// Write compression algorithm to compressed file
	WriteCompressionAlgorithm(algorithm, &compressedFile)
	
	// Encrypt compressed content if password is provided, else just compress
	err = encryption.Encrypt(&compressedFile.Content, password)
	if err != nil {
		return fmt.Errorf("encryption error: %w", err)
	}

	// Check if the output directory exists; create if not
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.MkdirAll(outputDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Check if the output file already exists; rename if necessary
	if _, err := os.Stat(outputDir + "/" + compressedFile.Name); err == nil {
		InvalidateFileName(&compressedFile, &outputDir)
	}

	output := outputDir + "/" + compressedFile.Name

	// Write compressed file to the output directory
	err = os.WriteFile(output, compressedFile.Content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write compressed file: %w", err)
	}

	//get the file size
	compressedFileInfo, err := os.Stat(output)
	if err != nil {
		return fmt.Errorf("failed to get compressed file info: %w", err)
	}

	compressedSize = compressedFileInfo.Size()

	// Calculate compression ratio
	compressionRatio := float64(compressedSize) / float64(originalSize)

	// Print compression statistics
	utils.ColorPrint(utils.GREEN, fmt.Sprintf("Compression complete: Original Size: %s, Compressed Size: %s, Compression Ratio: %.2f%%\n",
		utils.FileSize(originalSize), utils.FileSize(compressedSize), compressionRatio*100))

	return nil
}

func ReadAllFilesConcurrently(filenameStrs []string, files *[]utils.File, originalSize *int64) []error {
	// Use a wait group to synchronize goroutines
	var wg sync.WaitGroup
	var errMutex sync.Mutex // Mutex to handle errors safely

	// Channel to receive errors from goroutines
	errChan := make(chan error, len(filenameStrs))
	// Process each input file or folder recursively
	for _, filename := range filenameStrs {
		wg.Add(1)
		file := filename
		go readFileFromDisk(file, files, originalSize, &wg, errChan)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChan)

	errors := make([]error, 0)

	// Check for errors from goroutines
	for err := range errChan {
		if err != nil {
			//print all errors
			errMutex.Lock()
			errors = append(errors, err)
			errMutex.Unlock()
		}
	}

	return errors
}

func readFileFromDisk(filePath string, files *[]utils.File, originalSize *int64, wg *sync.WaitGroup, errChan chan error) {

	defer wg.Done()

	// Check if the file or folder exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		errChan <- fmt.Errorf("file or folder does not exist: '%s'", filePath)
		return
	}

	utils.ColorPrint(utils.YELLOW, fmt.Sprintf("Compressing file (%s)\n", filePath))
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		errChan <- fmt.Errorf("failed to read file '%s': %w", filePath, err)
		return
	}

	// Store file information (name and content)
	*files = append(*files, utils.File{
		Name:    filepath.Base(filePath),
		Content: content,
	})

	// Increment original size
	atomic.AddInt64(originalSize, info.Size())
}

// Decompress decompresses the specified compressed file into individual files or folders.
func Decompress(compressedFilename string, outputDir string, password string) error {

	// Set default output directory if not provided
	if outputDir == "" {
		outputDir = filepath.Dir(compressedFilename)
	}

	compressedContent := make([]byte, 0)
	var err error

	// Read compressed file content
	compressedContent, err = os.ReadFile(compressedFilename)
	if err != nil {
		return fmt.Errorf("failed to read compressed file %s: %w", compressedFilename, err)
	}

	// Decrypt compressed content if password is provided
	err = encryption.Decrypt(&compressedContent, password)

	if err != nil {
		return fmt.Errorf("decryption error: %w", err)
	}

	file := utils.File{
		Name:    filepath.Base(compressedFilename),
		Content: compressedContent,
	}

	// Read compression algorithm
	algorithm, err := ReadCompressionAlgorithm(&file)
	if err != nil {
		return fmt.Errorf("failed to read compression algorithm: %w", err)
	}

	var files []utils.File

	switch algorithm {
	case utils.HUFFMAN:
		files, err = hc.Unzip(file)
	case utils.ARITHMETIC:
		files, err = ar.Unzip(file)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}
	// Write decompressed files to disk
	err = writeAllFilesToDiskConcurrently(files, outputDir)
	if err != nil {
		return err
	}

	utils.ColorPrint(utils.GREEN, "Decompression complete\n")

	return nil
}

func writeAllFilesToDiskConcurrently(files []utils.File, outputDir string) error {

	// Use a wait group to synchronize goroutines
	var wg sync.WaitGroup
	var errMutex sync.Mutex // Mutex to handle errors safely

	// Channel to receive errors from goroutines
	errChan := make(chan error, len(files))

	// Process each decompressed file
	for _, file := range files {
		wg.Add(1)
		go writeFileToDisk(file, outputDir, &wg, errChan)
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

func writeFileToDisk(file utils.File, outputDir string, wg *sync.WaitGroup, errChan chan error) {

	defer wg.Done()

	// Create directories if they don't exist
	outputPath := filepath.Join(outputDir, filepath.Dir(file.Name))
	err := os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		errChan <- fmt.Errorf("failed to create directory %s: %w", outputPath, err)
		return
	}

	// Check if the file already exists, rename it with file_N
	if _, err := os.Stat(filepath.Join(outputDir, file.Name)); err == nil {
		InvalidateFileName(&file, &outputDir)
	}

	// Write decompressed file content
	err = os.WriteFile(filepath.Join(outputDir, file.Name), file.Content, 0644)
	if err != nil {
		errChan <- fmt.Errorf("failed to write decompressed file %s: %w", file.Name, err)
		return
	}
	utils.ColorPrint(utils.YELLOW, fmt.Sprintf("Decompressed file: %s\n", file.Name))
}
