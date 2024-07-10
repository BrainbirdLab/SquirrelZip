package compressor

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"file-compressor/encryption"
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
	var originalSize int64
	var compressedSize int64

	// Use a wait group to synchronize goroutines
	var wg sync.WaitGroup
	var errMutex sync.Mutex // Mutex to handle errors safely

	// Channel to receive errors from goroutines
	errChan := make(chan error, len(filenameStrs))
	// Process each input file or folder recursively
	for _, filename := range filenameStrs {
		wg.Add(1)
		file := filename
		go handleFileFolder(file, &files, &originalSize, &wg, errChan)
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
	compressedFile, err := Zip(files)

	if err != nil {
		return err
	}

	// Encrypt compressed content if password is provided
	if password != nil && *password != "" {
		compressedFile.Content, err = encryption.Encrypt(compressedFile.Content, *password)
		if err != nil {
			return fmt.Errorf("encryption error: %w", err)
		}
	}

	// Check if the output directory exists; create if not
	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		err := os.MkdirAll(*outputDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Check if the output file already exists; rename if necessary
	if _, err := os.Stat(*outputDir + "/" + compressedFile.Name); err == nil {
		InvalidateFileName(&compressedFile, outputDir)
	}

	// Write compressed file to the output directory
	err = os.WriteFile(*outputDir+"/"+compressedFile.Name, compressedFile.Content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write compressed file: %w", err)
	}

	//get the file size
	compressedFileInfo, err := os.Stat(*outputDir + "/" + compressedFile.Name)
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

func handleFileFolder(filePath string, files *[]utils.File, originalSize *int64, wg *sync.WaitGroup, errChan chan error) {
	
	defer wg.Done()

	// Check if the file or folder exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		errChan <- fmt.Errorf("file or folder does not exist: %s", filePath)
	}

	if info.IsDir() {
		utils.ColorPrint(utils.YELLOW, fmt.Sprintf("Compressing folder (%s)\n", filePath))
		err := compressFolderRecursive(filePath, files, originalSize)
		if err != nil {
			errChan <- err
		}
	} else {
		utils.ColorPrint(utils.YELLOW, fmt.Sprintf("Compressing file (%s)\n", filePath))
		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			errChan <- fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// Store file information (name and content)
		*files = append(*files, utils.File{
			Name:    path.Base(filePath),
			Content: content,
		})

		// Increment original size
		atomic.AddInt64(originalSize, info.Size())
	}
}

// Function to recursively compress a folder and its contents
func compressFolderRecursive(folderPath string, files *[]utils.File, originalSize *int64) error {
	// Traverse the folder contents
	err := filepath.Walk(folderPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			utils.ColorPrint(utils.YELLOW, fmt.Sprintf("Compressing file (%s)\n", filePath))
			// Read file content
			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, err)
			}

			// Store file information (relative path and content)
			relativePath, err := filepath.Rel(folderPath, filePath)
			if err != nil {
				return fmt.Errorf("failed to get relative path for file %s: %w", filePath, err)
			}

			*files = append(*files, utils.File{
				Name:    filepath.ToSlash(relativePath), // Store relative path
				Content: content,
			})

			// Increment original size and compressed size
			atomic.AddInt64(originalSize, info.Size())
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
		compressedContent, err = encryption.Decrypt(compressedContent, *password)
		if err != nil {
			return fmt.Errorf("decryption error: %w", err)
		}
	}

	// Decompress file using Huffman coding
	files, err := Unzip(utils.File{
		Name:    path.Base(filenameStrs[0]),
		Content: compressedContent,
	})

	if err != nil {
		return err
	}

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

			// Check if the file already exists, rename it with file_N
			if _, err := os.Stat(filepath.Join(*outputDir, file.Name)); err == nil {
				InvalidateFileName(&file, outputDir)
			}

			// Write decompressed file content
			err = os.WriteFile(filepath.Join(*outputDir, file.Name), file.Content, 0644)
			if err != nil {
				errChan <- fmt.Errorf("failed to write decompressed file %s: %w", file.Name, err)
				return
			}
			utils.ColorPrint(utils.YELLOW, fmt.Sprintf("Decompressed file: %s\n", file.Name))
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

	utils.ColorPrint(utils.GREEN, "Decompression complete\n")

	return nil
}

func InvalidateFileName(file *utils.File, outputDir *string) {
	fileExt := path.Ext(file.Name)
	//extract the file name without the extension
	filename := path.Base(file.Name)
	filename = strings.TrimSuffix(filename, fileExt)

	count := 1
	for {
		if _, err := os.Stat(filepath.Join(*outputDir, filename + fmt.Sprintf("_%d%s", count, fileExt))); os.IsNotExist(err) {
			utils.ColorPrint(utils.PURPLE, fmt.Sprintf("File %s already exists, renaming to %s\n", file.Name, filename+fmt.Sprintf("_%d%s", count, fileExt)))
			file.Name = filename + fmt.Sprintf("_%d%s", count, fileExt)
			break
		}
		count++
	}
}
