package compressor

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"path/filepath"

	"file-compressor/utils"
	"file-compressor/encryption"
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
		compressedFile.Content, err = encryption.Encrypt(compressedFile.Content, *password)
		if err != nil {
			return fmt.Errorf("encryption error: %w", err)
		}
	}

	//check if the output file and directory exists
	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		err := os.MkdirAll(*outputDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// if the output file already exists, rename it with file(N).bin
	if _, err := os.Stat(*outputDir + "/" + compressedFile.Name + ".bin"); err == nil {
		count := 1
		for {
			if _, err := os.Stat(*outputDir + "/" + compressedFile.Name + fmt.Sprintf("(%d).bin", count)); os.IsNotExist(err) {
				compressedFile.Name = compressedFile.Name + fmt.Sprintf("(%d).bin", count)
				break
			}
			count++
		}
	} else if os.IsNotExist(err) {
		compressedFile.Name = compressedFile.Name + ".bin"
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
		compressedContent, err = encryption.Decrypt(compressedContent, *password)
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