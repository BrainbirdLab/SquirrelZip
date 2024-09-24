package compressor

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"file-compressor/compressor/hfc"
	"file-compressor/constants"
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
func Compress(filenameStrs []string, outputDir, algorithm string) (string, utils.FilesRatio, error) {

	fileMeta := utils.FilesRatio{}
	//check if files exist
	for _, filenameStr := range filenameStrs {
		if _, err := os.Stat(filenameStr); os.IsNotExist(err) {
			return "", fileMeta, fmt.Errorf("file '%s' does not exist", filenameStr)
		}
	}

	// Check if the compression algorithm is supported
	err := CheckCompressionAlgorithm(algorithm)
	if err != nil {
		return "", fileMeta, err
	}

	// Set default output directory if not provided
	setOutputDir(&outputDir, filenameStrs[0])

	// Check if the output directory exists, create it if it doesn't
	if err := utils.MakeOutputDir(outputDir); err != nil {
		return "", fileMeta, err
	}


	fileName := filenameStrs[0]
	ext := filepath.Ext(fileName)
	fileName = strings.TrimSuffix(fileName, ext)
	fileName = utils.InvalidateFileName(fileName, outputDir)

	compressedFile, err := os.Create(fileName)
	if err != nil {
		return "", fileMeta, fmt.Errorf(constants.ERROR_COMPRESS, err)
	}
	
	defer compressedFile.Close()

	originalSize, err := ReadAndCompressFiles(filenameStrs, compressedFile, algorithm)
	if err != nil {
		return "", fileMeta, err
	}

	compressedStat, err := os.Stat(fileName)
	if err != nil {
		return "", fileMeta, fmt.Errorf(constants.FILE_STAT_ERROR, err)
	}

	fileMeta = utils.NewFilesRatio(originalSize, uint64(compressedStat.Size()))

	return fileName, fileMeta, err
}

// ReadAndCompressFiles reads all files concurrently and writes the compressed data to the output.
func ReadAndCompressFiles(filenameStrs []string, output io.Writer, algorithm string) (uint64, error) {

	var err error

	fileDataArr := []utils.FileData{}

	originalSize := uint64(0)

	for _, filenameStr := range filenameStrs {
		// Get the file info
		fileInfo, err := os.Stat(filenameStr)
		if err != nil {
			return 0, fmt.Errorf("failed to get file info: %v", err)
		}

		fmt.Printf("File info: %s, size: %d\n", filenameStr, fileInfo.Size())

		originalSize += uint64(fileInfo.Size())

		// Check if the file is a directory
		if fileInfo.IsDir() {
			if err := walkDir(filenameStr, &fileDataArr); err != nil {
				return 0, err
			}
		} else {
			file, err := os.Open(filenameStr)
			if err != nil {
				return 0, fmt.Errorf(constants.FILE_OPEN_ERROR, err)
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
		return 0, err
	}

	switch utils.Algorithm(algorithm) {
	case utils.HUFFMAN:
		err = hfc.Zip(fileDataArr, output)
	}

	if err != nil {
		return 0, fmt.Errorf(constants.ERROR_COMPRESS, err)
	}

	return originalSize, nil
}

func writeAlgorithm(output io.Writer, algorithm string) error {
	// write the length of compression algorithm
	if err := binary.Write(output, binary.LittleEndian, uint8(len(algorithm))); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	// write the compression algorithm
	if err := binary.Write(output, binary.LittleEndian, []byte(algorithm)); err != nil {
		return fmt.Errorf(constants.BUFFER_WRITE_ERROR, err)
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
			return nil, fmt.Errorf(constants.ERROR_DECOMPRESS, err)
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
func Decompress(compressedFilePath, outputDir string) ([]string, error) {

	outputFiles := make([]string, 0)
	// check if the compressed file exists
	if _, err := os.Stat(compressedFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("compressed file '%s' does not exist", compressedFilePath)
	}

	// decrypt the compressed file first
	compressedFile, err := os.Open(compressedFilePath)
	if err != nil {
		return outputFiles, fmt.Errorf(constants.FILE_OPEN_ERROR, err)
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
		return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
	}

	// read the compression algorithm
	algo := make([]byte, algoLen)
	if err = binary.Read(compressedFile, binary.LittleEndian, &algo); err != nil {
		return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
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
			return fmt.Errorf(constants.FILE_OPEN_ERROR, err)
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