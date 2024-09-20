package compressor

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"file-compressor/compressor/hfc"
	"file-compressor/utils"
	ec "file-compressor/errorConstants"
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

	// Set default output directory as the directory of the first file if not provided
	setOutputDir(&outputDir, filenameStrs[0])


	// Check if the output directory exists, create it if it doesn't
	if err := utils.MakeOutputDir(outputDir); err != nil {
		return "", 0, err
	}

	var originalSize int64

	//var compressedSize int64
	outputFilename := []byte("compressed.sq")
	filePath := filepath.Join(outputDir, string(outputFilename))

	// Create a writer to write the compressed data
	outputFile, err := os.Create(filePath)
	if err != nil {
		return "", 0, fmt.Errorf(ec.ERROR_COMPRESS, err)
	}

	defer outputFile.Close()

	err = ReadAndCompressFiles(filenameStrs, outputFile, &originalSize, algorithm)

	return filePath, originalSize, err
}

// ReadAndCompressFiles reads all files concurrently and writes the compressed data to the output.
func ReadAndCompressFiles(filenameStrs []string, output io.Writer, originalSize *int64, algorithm string) error {

	var err error

	fileDataArr := []utils.FileData{}

	for _, filenameStr := range filenameStrs {
		fileInfo, err := os.Stat(filenameStr)
		if err != nil {
			return fmt.Errorf("failed to get file info: %v", err)
		}

		*originalSize += fileInfo.Size()

		// Check if the file is a directory
		if fileInfo.IsDir() {
			if err := walkDir(filenameStr, &fileDataArr); err != nil {
				return err
			}
		} else {
			file, err := os.Open(filenameStr)
			if err != nil {
				return fmt.Errorf(ec.FILE_OPEN_ERROR, err)
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

	// Write the compression algorithm to the output
	if err := writeAlgorithm(output, algorithm); err != nil {
		return err
	}

	switch utils.Algorithm(algorithm) {
	case utils.HUFFMAN:
		err = hfc.Zip(fileDataArr, output)
	}
	// Compress the files
	
	if err != nil {
		return fmt.Errorf(ec.ERROR_COMPRESS, err)
	}

	return err
}

func writeAlgorithm(output io.Writer, algorithm string) error {
	// write the length of compression algorithm
	if err := binary.Write(output, binary.LittleEndian, uint8(len(algorithm))); err != nil {
		return fmt.Errorf(ec.FILE_WRITE_ERROR, err)
	}

	// write the compression algorithm
	if err := binary.Write(output, binary.LittleEndian, []byte(algorithm)); err != nil {
		return fmt.Errorf(ec.BUFFER_WRITE_ERROR, err)
	}

	return nil
}

func WriteAndDecompressFiles(compressedFile io.Reader, outputDir string, algorithm []byte) ([]string, error) {

	var fileNames []string
	var err error

	switch utils.Algorithm(algorithm) {
	case utils.HUFFMAN:
		// Decompress the file
		fileNames, err = hfc.Unzip(compressedFile, outputDir)
		if err != nil {
			return nil, fmt.Errorf(ec.ERROR_DECOMPRESS, err)
		}
	}

	return fileNames, nil
}

// Decompress decompresses the specified compressed file.
/*
compressedFilePath: the path to the compressed file
outputDir: the directory to save the decompressed files
password: the password to decrypt the compressed file
algorithm: the compression algorithm used

returns the list of decompressed files and an error if any
*/
func Decompress(compressedFilePath, outputDir, password string) ([]string, error) {
	outputFiles := make([]string, 0)

	// Open the compressed file
	compressedFile, err := os.Open(compressedFilePath)
	if err != nil {
		return nil, fmt.Errorf(ec.FILE_OPEN_ERROR, err)
	}

	defer compressedFile.Close()

	// Read the compression algorithm
	algorithm, err := readAlgorithm(compressedFile)
	if err != nil {
		return outputFiles, err
	}

	// Check if the compression algorithm is supported
	err = CheckCompressionAlgorithm(string(algorithm))
	if err != nil {
		return outputFiles, err
	}

	setOutputDir(&outputDir, compressedFilePath)

	// Check if the output directory exists
	if err := utils.MakeOutputDir(outputDir); err != nil {
		return nil, err
	}

	// Decompress the file
	fileNames, err := WriteAndDecompressFiles(compressedFile, outputDir, algorithm)
	if err != nil {
		return outputFiles, err
	}

	return fileNames, nil
}

func readAlgorithm(compressedFile io.Reader) ([]byte, error) {
	var algoLen uint8
	var err error

	// read the length of compression algorithm
	if err = binary.Read(compressedFile, binary.LittleEndian, &algoLen); err != nil {
		return nil, fmt.Errorf(ec.FILE_READ_ERROR, err)
	}

	// read the compression algorithm
	algo := make([]byte, algoLen)
	if err = binary.Read(compressedFile, binary.LittleEndian, &algo); err != nil {
		return nil, fmt.Errorf(ec.FILE_READ_ERROR, err)
	}

	return algo, nil
}

func setOutputDir(outputDir *string, firstFilename string) {
	// Set default output directory if not provided
	if *outputDir == "" {
		*outputDir = filepath.Dir(firstFilename)
	}
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
			return fmt.Errorf(ec.FILE_OPEN_ERROR, err)
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