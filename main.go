package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"file-compressor/compressor"
	"file-compressor/utils"
)

func main() {
	// Command-line arguments
	inputFiles := flag.String("i", "", "Input file(s) to process (comma-separated for multiple files)")
	password := flag.String("p", "", "Password for encryption/decryption (optional)")
	decrypt := flag.Bool("d", false, "Decrypt the input file")
	flag.Parse()

	// Validate input
	if *inputFiles == "" {
		utils.ColorPrint(utils.RED, "At least one input file is required.\n")
		fmt.Println("Usage: exe -i <input_file(s)> [-d] -p <password>")
		flag.Usage()
		os.Exit(1)
	}

	// Split comma-separated input files
	files := strings.Split(*inputFiles, ",")

	// Process each input file
	for _, inputFile := range files {
		// Determine output file name based on input file name and flag
		outputFile := determineOutputFileName(inputFile, *decrypt)

		// Perform decryption or compression based on flags
		if *decrypt {
			// Decrypting
			performDecryption(inputFile, outputFile, *password)
		} else {
			// Compressing
			performCompression(inputFile, outputFile, *password)
		}
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
		outputFile = inputFile + ".pcd" // Add .pcd extension for compressed file
	}
	return outputFile
}

func performDecryption(inputFile, outputFile, password string) {
	// Read compressed file with encryption flag
	data, codes, encrypted, err := compressor.ReadCompressedFile(inputFile)
	if err != nil {
		utils.ColorPrint(utils.RED, fmt.Sprintf("Error reading compressed file %s: %s\n", inputFile, err.Error()))
		return
	}

	// If the file is not encrypted
	if !encrypted {
		utils.ColorPrint(utils.RED, fmt.Sprintf("File %s is not encrypted. Cannot decrypt.\n", inputFile))
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
	decompressed := compressor.Decompress(decrypted, codes)

	// Write decompressed data to output file
	if err := os.WriteFile(outputFile, decompressed, 0644); err != nil {
		utils.ColorPrint(utils.RED, fmt.Sprintf("Error writing decompressed file %s: %s\n", outputFile, err.Error()))
		return
	}

	utils.ColorPrint(utils.GREEN, fmt.Sprintf("File %s decrypted successfully. Output file: %s\n", inputFile, outputFile))
}

func performCompression(inputFile, outputFile, password string) {
	// Read input file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		utils.ColorPrint(utils.RED, fmt.Sprintf("Error reading input file %s: %s\n", inputFile, err.Error()))
		return
	}

	// Compress data
	compressed, codes := compressor.Compress(data)

	// Encrypt if password is provided
	if password != "" {
		compressed = utils.Encrypt(compressed, password)
	}

	// Write compressed file with encryption flag
	if err := compressor.WriteCompressedFile(outputFile, compressed, codes, password != ""); err != nil {
		utils.ColorPrint(utils.RED, fmt.Sprintf("Error writing compressed file %s: %s\n", outputFile, err.Error()))
		return
	}

	utils.ColorPrint(utils.GREEN, fmt.Sprintf("File %s compressed successfully. Output file: %s\n", inputFile, outputFile))
}
