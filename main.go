package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"file-compressor/compressor"
	"file-compressor/utils"
)

func main() {
	// Command-line arguments
	inputFiles := flag.String("i", "", "Input file(s) to process (comma-separated for multiple files)")
	password := flag.String("p", "", "Password for encryption/decryption (optional)")
	decrypt := flag.Bool("d", false, "Decrypt the input file(s)")
	combine := flag.Bool("c", false, "Combine all files into a single compressed file")
	flag.Parse()

	// Validate input
	if *inputFiles == "" {
		utils.ColorPrint(utils.RED, "At least one input file is required.\n")
		fmt.Println("Usage: exe -i <input_file(s)> [-d] [-c] -p <password>")
		flag.Usage()
		os.Exit(1)
	}

	// Split comma-separated input files
	files := strings.Split(*inputFiles, ",")

	// Perform decryption if decrypt flag is set
	if *decrypt {
		performConcurrentDecryption(files, *password)
		return
	}

	// Perform compression or combined compression based on flags
	if *combine {
		// Combine all files into a single compressed file
		performCombinedCompression(files, *password)
	} else {
		// Compress each file individually
		performIndividualCompression(files, *password)
	}
}

func determineOutputFileName(inputFile string, decrypt bool) string {
	var outputFile string
	if decrypt {
		// Remove .pcd extension if present
		base := filepath.Base(inputFile)
		ext := filepath.Ext(base)
		if ext == ".pcd" {
			outputFile = strings.TrimSuffix(base, ext)
		} else {
			// Handle error for incorrect extension during decryption
			utils.ColorPrint(utils.RED, "Error: Input file must have .pcd extension for decryption.\n")
			os.Exit(1)
		}
	} else {
		// Use a generic .pcd extension without exposing the original file type
		outputFile = strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile)) + ".pcd"
		dir := filepath.Dir(inputFile)
		outputFile = filenameClearance(dir + "/" + outputFile)
	}
	return outputFile
}

func performConcurrentDecryption(files []string, password string) {
	// Create a wait group to manage go routines
	var wg sync.WaitGroup

	// Channel to communicate errors from go routines
	errCh := make(chan error, len(files))

	// Launch go routines for decryption
	for _, inputFile := range files {
		wg.Add(1)
		go func(inputFile string) {
			defer wg.Done()

			outputFile := determineOutputFileName(inputFile, true)
			performDecryption(inputFile, outputFile, password)
		}(inputFile)
	}

	// Wait for all go routines to finish
	wg.Wait()
	close(errCh)

	// Check for any errors from go routines
	for err := range errCh {
		utils.ColorPrint(utils.RED, fmt.Sprintf("%v\n", err))
	}
}

func performDecryption(inputFile, outputFile, password string) {
	// Read compressed file with encryption flag and original extension
	data, codes, encrypted, originalExt, err := compressor.ReadCompressedFile(inputFile)
	if err != nil {
		utils.ColorPrint(utils.RED, fmt.Sprintf("Error reading compressed file %s: %s\n", inputFile, err.Error()))
		return
	}

	// Decompress data if the file is not encrypted
	if !encrypted {
		decompressData(data, codes, inputFile, outputFile, originalExt)
		return
	}

	// If password is provided, decrypt the data
	if password == "" {
		utils.ColorPrint(utils.RED, fmt.Sprintf("Password is required for decryption of file %s.\n", inputFile))
		return
	}

	// Decrypt data
	decrypted, err := utils.Decrypt(data, password)
	if err != nil {
		utils.ColorPrint(utils.RED, fmt.Sprintf("Error decrypting file %s: %s\n", inputFile, err.Error()))
		return
	}

	// Decompress data
	decompressData(decrypted, codes, inputFile, outputFile, originalExt)
}

func decompressData(data []byte, codes map[rune]string, inputFile, outputFile, originalExt string) {
	decompressed, err := compressor.Decompress(data, codes)
	if err != nil {
		utils.ColorPrint(utils.RED, fmt.Sprintf("Error decompressing file %s: %s\n", inputFile, err.Error()))
		return
	}

	// Write decompressed data to output file with the original extension
	outputFile = outputFile + originalExt
	if err := os.WriteFile(outputFile, decompressed, 0644); err != nil {
		utils.ColorPrint(utils.RED, fmt.Sprintf("Error writing decompressed file %s: %s\n", outputFile, err.Error()))
		return
	}

	utils.ColorPrint(utils.GREEN, fmt.Sprintf("File %s processed successfully. Output file: %s\n", inputFile, outputFile))
}

func filenameClearance(file string) string {
	//if exist, rename
	if _, err := os.Stat(file); err == nil {
		i := 1
		for {
			file = fmt.Sprintf("%s_%d", file, i)
			if _, err := os.Stat(file); err != nil {
				break
			}
			i++
		}
	}

	return file
}


func performCombinedCompression(files []string, password string) {

	dir := filepath.Dir(files[0])

	// Prepare combined output file name
	outputFile := dir + "/files.pcd"

	//if file exists then rename with number
	outputFile = filenameClearance(outputFile)

	// Create a wait group to manage go routines
	var wg sync.WaitGroup

	// Channel to communicate errors from go routines
	errCh := make(chan error, len(files))

	// Launch go routines for compression
	for _, inputFile := range files {
		wg.Add(1)
		go func(inputFile string) {
			defer wg.Done()

			// Read input file
			data, err := os.ReadFile(inputFile)
			if err != nil {
				errCh <- fmt.Errorf("error reading input file %s: %w", inputFile, err)
				return
			}

			// Compress data
			compressed, codes := compressor.Compress(data)

			// Encrypt if password is provided
			if password != "" {
				compressed = utils.Encrypt(compressed, password)
			}

			// Write compressed data to combined file with encryption flag and original file extension
			originalExt := filepath.Ext(inputFile)
			if err := compressor.WriteCompressedFile(outputFile, compressed, codes, password != "", originalExt); err != nil {
				errCh <- fmt.Errorf("error writing compressed file %s: %w", outputFile, err)
				return
			}

			utils.ColorPrint(utils.GREEN, fmt.Sprintf("File %s compressed successfully and added to %s\n", inputFile, outputFile))
		}(inputFile)
	}

	// Wait for all go routines to finish
	wg.Wait()
	close(errCh)

	// Check for any errors from go routines
	for err := range errCh {
		utils.ColorPrint(utils.RED, fmt.Sprintf("%v\n", err))
	}
}

func performIndividualCompression(files []string, password string) {
	// Create a wait group to manage go routines
	var wg sync.WaitGroup

	// Channel to communicate errors from go routines
	errCh := make(chan error, len(files))

	// Launch go routines for compression
	for _, inputFile := range files {
		wg.Add(1)
		go func(inputFile string) {
			defer wg.Done()

			// Determine output file name based on input file name and flag
			outputFile := determineOutputFileName(inputFile, false)

			// Read input file
			data, err := os.ReadFile(inputFile)
			if err != nil {
				errCh <- fmt.Errorf("error reading input file %s: %w", inputFile, err)
				return
			}

			// Compress data
			compressed, codes := compressor.Compress(data)

			// Encrypt if password is provided
			if password != "" {
				compressed = utils.Encrypt(compressed, password)
			}

			// Write compressed file with encryption flag and original file extension
			originalExt := filepath.Ext(inputFile)
			if err := compressor.WriteCompressedFile(outputFile, compressed, codes, password != "", originalExt); err != nil {
				errCh <- fmt.Errorf("error writing compressed file %s: %w", outputFile, err)
				return
			}

			utils.ColorPrint(utils.GREEN, fmt.Sprintf("File %s compressed successfully. Output file: %s\n", inputFile, outputFile))
		}(inputFile)
	}

	// Wait for all go routines to finish
	wg.Wait()
	close(errCh)

	// Check for any errors from go routines
	for err := range errCh {
		utils.ColorPrint(utils.RED, fmt.Sprintf("%v\n", err))
	}
}
