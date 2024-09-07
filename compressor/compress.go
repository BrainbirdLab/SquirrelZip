package compressor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"file-compressor/compressor/hfc"
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
/*
filenameStrs: the list of files or folders to compress
outputDir: the directory to save the compressed file
password: the password to encrypt the compressed file
algorithm: the compression algorithm to use

returns the path to the compressed file, the original size of the files, and an error if any
*/
func Compress(filenameStrs []string, outputDir, password, algorithm string) (string, int64, error) {

	// Check if the compression algorithm is supported
	err := CheckCompressionAlgorithm(algorithm)
	if err != nil {
		return "", 0, err
	}

	setOutputDir(&outputDir, filenameStrs[0])

	fmt.Printf("Output directory: %s\n", outputDir)

	// Check if the output directory exists
	if err := checkOutputDir(outputDir); err != nil {
		return "", 0, err
	}

	var originalSize int64
	//var compressedSize int64
	outputFilename := []byte("compressed.sq")
	filePath := filepath.Join(outputDir, string(outputFilename))

	// Create a writer to write the compressed data
	outputFile, err := os.Create(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create compressed file: %v", err)
	}
	defer outputFile.Close()

	// Read all files concurrently
	errs := ReadAndCompressFiles(filenameStrs, outputFile, &originalSize)

	// Check for errors from goroutines
	if len(errs) > 0 {
		for _, err := range errs {
			utils.ColorPrint(utils.RED, fmt.Sprintf("Error: %v\n", err))
		}
		os.Exit(1)
	}

	return filePath, originalSize, nil
}

// ReadAndCompressFiles reads all files concurrently and writes the compressed data to the output.
func ReadAndCompressFiles(filenameStrs []string, output io.Writer, originalSize *int64) []error {
	errs := make([]error, 0)

	fileDataArr := []utils.FileData{}

	for _, filenameStr := range filenameStrs {
		fileInfo, err := os.Stat(filenameStr)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get file info: %v", err))
			continue
		}

		*originalSize += fileInfo.Size()

		// Check if the file is a directory
		if fileInfo.IsDir() {
			errs = append(errs, walkDir(filenameStr, &fileDataArr))
		} else {
			file, err := os.Open(filenameStr)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to open file: %v", err))
				continue
			}

			defer file.Close()

			fileData := utils.FileData{
				Name: filenameStr,
				Size: fileInfo.Size(),
				Reader: file,
			}

			fileDataArr = append(fileDataArr, fileData)
		}
	}

	// Compress the files
	err := hfc.Zip(fileDataArr, output)
	
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to compress files: %v", err))
	}

	return errs
}

func WriteAndDecompressFiles(inputFileName string, outputDir string) ([]string, error) {
	errs := make([]error, 0)

	// Open the compressed file
	compressedFile, err := os.Open(inputFileName)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to open compressed file: %v", err))
		return nil, err
	}

	defer compressedFile.Close()

	// Decompress the file
	fileNames, err := hfc.Unzip(compressedFile, outputDir)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to decompress file: %v", err))
	}

	return fileNames, errs[0]
}

// Decompress decompresses the specified compressed file.
/*
compressedFilePath: the path to the compressed file
outputDir: the directory to save the decompressed files
password: the password to decrypt the compressed file
algorithm: the compression algorithm used

returns the list of decompressed files and an error if any
*/
func Decompress(compressedFilePath, outputDir, password, algorithm string) ([]string, error) {
	outputFiles := make([]string, 0)

	// Check if the compression algorithm is supported
	err := CheckCompressionAlgorithm(algorithm)
	if err != nil {
		return outputFiles, err
	}

	setOutputDir(&outputDir, compressedFilePath)

	// Check if the output directory exists
	if err := checkOutputDir(outputDir); err != nil {
		return nil, err
	}

	// Decompress the file
	fileNames, err := WriteAndDecompressFiles(compressedFilePath, outputDir)
	if err != nil {
		return outputFiles, err
	}

	return fileNames, nil
}

func setOutputDir(outputDir *string, firstFilename string) {
	// Set default output directory if not provided
	if *outputDir == "" {
		*outputDir = filepath.Dir(firstFilename)
	}
}

func checkOutputDir(outputDir string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.Mkdir(outputDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}
	
	return nil
}

// walkDir walks the directory and reads all files in the directory.
func walkDir(filenameStr string, fileDataArr *[]utils.FileData) error {
	err := filepath.Walk(filenameStr, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk path: %v", err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		defer file.Close()

		fileData := utils.FileData{
			Name: path,
			Size: info.Size(),
			Reader: file,
		}

		*fileDataArr = append(*fileDataArr, fileData)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %v", err)
	}

	return nil
}