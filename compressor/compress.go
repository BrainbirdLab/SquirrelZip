package compressor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"file-compressor/utils"
)

func CheckCompressionAlgorithm(algo string) error {
	switch utils.Algorithm(algo) {
	case utils.HUFFMAN, utils.ARITHMETIC:
		return nil
	default:
		return fmt.Errorf("unsupported compression algorithm: %v", algo)
	}
}


// Compress compresses the specified files or folders into a single compressed file.
func Compress(filenameStrs []string, outputDir string, password string, algorithm string, tempPath string) error {
	// Set default output directory if not provided
	if outputDir == "" {
		outputDir = filepath.Dir(filenameStrs[0])
	}

	fmt.Printf("Output directory: %s\n", outputDir)

	// if tempPath does not exist, create it
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		err := os.Mkdir(tempPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create temporary directory: %w", err)
		}
	}

	// Prepare to store files' content
	tempFile, err := os.CreateTemp(tempPath, "temp-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tempFile.Close()


	var originalSize int64
	//var compressedSize int64

	// Read all files concurrently
	errs := ReadAllFilesConcurrently(filenameStrs, tempFile, &originalSize)

	// Check for errors from goroutines
	if len(errs) > 0 {
		for _, err := range errs {
			utils.ColorPrint(utils.RED, fmt.Sprintf("Error: %v\n", err))
		}
		os.Exit(1)
	}


	// Check if the compression algorithm is supported
	err = CheckCompressionAlgorithm(algorithm)
	if err != nil {
		return err
	}

	return nil
}

func ReadAllFilesConcurrently(filenameStrs []string, sharedFile io.Writer, originalSize *int64) []error {
	// Use a wait group to synchronize goroutines
	var wg sync.WaitGroup
	var errMutex sync.Mutex // Mutex to handle errors safely

	// Channel to receive errors from goroutines
	errChan := make(chan error, len(filenameStrs))
	// Process each input file or folder recursively
	for _, filename := range filenameStrs {
		wg.Add(1)
		file := filename
		go readFileFromDisk(file, sharedFile, originalSize, &wg, errChan)
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

func readFileFromDisk(filePath string, sharedFile io.Writer, originalSize *int64, wg *sync.WaitGroup, errChan chan error) {

	defer wg.Done()

	// Check if the file or folder exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		errChan <- fmt.Errorf("file or folder does not exist: '%s'", filePath)
		return
	}

	utils.ColorPrint(utils.YELLOW, fmt.Sprintf("Compressing file (%s)\n", filePath))

	// write the file content to the shared file
	err = writeFileContentToSharedFile(filePath, sharedFile)
	if err != nil {
		errChan <- fmt.Errorf("failed to write file content: %w", err)
		return
	}

	// Increment original size
	atomic.AddInt64(originalSize, info.Size())
}

func writeFileContentToSharedFile(filePath string, sharedFile io.Writer) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	// Write the file information first
	// write filename length
	_, err = sharedFile.Write([]byte{byte(len(filePath))})
	if err != nil {
		return fmt.Errorf("failed to write filename length: %w", err)
	}
	// write filename
	_, err = sharedFile.Write([]byte(filePath))
	if err != nil {
		return fmt.Errorf("failed to write filename: %w", err)
	}
	// Read the file content
	_, err = io.Copy(sharedFile, file)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}
	return nil
}