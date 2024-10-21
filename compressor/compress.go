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


// Compress compresses a list of files using the specified compression algorithm and saves the compressed file to the output directory.
// 
// Parameters:
// - filenameStrs: A slice of strings containing the paths of the files to be compressed.
// - outputDir: A string specifying the directory where the compressed file will be saved. If not provided, a default directory will be used.
// - algorithm: A string specifying the compression algorithm to be used.
//
// Returns:
// - A string representing the path of the compressed file.
// - A utils.FilesRatio struct containing the ratio of the original size to the compressed size.
// - An error if any issues occur during the compression process.
//
// The function performs the following steps:
// 1. Checks if the files in filenameStrs exist.
// 2. Verifies if the specified compression algorithm is supported.
// 3. Sets the default output directory if not provided.
// 4. Ensures the output directory exists, creating it if necessary.
// 5. Generates a valid name for the compressed file.
// 6. Creates the compressed file in the output directory.
// 7. Reads and compresses the input files using the specified algorithm.
// 8. Calculates the size ratio between the original and compressed files.
// 9. Returns the path of the compressed file, the size ratio, and any error encountered.
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
	fileName = filepath.Base(fileName) + constants.COMPRESSED_FILE_EXT
	fileName = utils.InvalidateFileName(fileName, outputDir)

	compressedFileOutput, err := os.Create(fileName)
	if err != nil {
		return "", fileMeta, fmt.Errorf(constants.ERROR_COMPRESS, err)
	}
	
	defer compressedFileOutput.Close()

	originalSize, err := ReadAndCompressFiles(filenameStrs, compressedFileOutput, algorithm)
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


// ReadAndCompressFiles reads a list of files, compresses them using the specified algorithm,
// and writes the compressed data to the provided output writer.
//
// Parameters:
//   - filenameStrs: A slice of strings containing the file paths to be read and compressed.
//   - output: An io.Writer where the compressed data will be written.
//   - algorithm: A string specifying the compression algorithm to use.
//
// Returns:
//   - uint64: The total size of the original uncompressed files.
//   - error: An error if any occurs during the process.
//
// The function performs the following steps:
//   1. Iterates over the provided file paths.
//   2. Retrieves file information and checks if the file is a directory.
//   3. If the file is a directory, it recursively walks through the directory to gather file data.
//   4. If the file is not a directory, it opens the file and appends its data to a slice.
//   5. Writes the specified compression algorithm to the output.
//   6. Compresses the gathered file data using the specified algorithm and writes the compressed data to the output.
//
// Supported compression algorithms:
//   - utils.HUFFMAN: Uses Huffman coding for compression.
//
// Errors:
//   - Returns an error if any file cannot be opened, read, or if compression fails.
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

// writeAlgorithm writes the specified compression algorithm name to the provided writer.
// It first writes the length of the algorithm name as a single byte, followed by the algorithm name itself.
//
// Parameters:
//   - output: An io.Writer where the algorithm name will be written.
//   - algorithm: A string representing the name of the compression algorithm.
//
// Returns:
//   - error: An error if writing to the output fails, otherwise nil.
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


// WriteAndDecompressFiles reads a compressed file from the provided io.Reader,
// decompresses it using the specified algorithm, and writes the decompressed
// files to the given output directory.
//
// Parameters:
//   - compressedFile: an io.Reader from which the compressed file is read.
//   - outputDir: a string specifying the directory where decompressed files will be written.
//   - algorithm: a byte slice indicating the decompression algorithm to use.
//
// Returns:
//   - A slice of strings containing the names of the decompressed files.
//   - An error if the decompression process fails.
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


// Decompress extracts files from a compressed archive.
//
// Parameters:
//   - compressedFilePath: The path to the compressed file to be decompressed.
//   - outputDir: The directory where the decompressed files will be stored.
//
// Returns:
//   - A slice of strings containing the names of the decompressed files.
//   - An error if any issue occurs during the decompression process.
//
// The function performs the following steps:
//   1. Checks if the compressed file exists.
//   2. Opens the compressed file.
//   3. Reads the compression algorithm used.
//   4. Verifies if the compression algorithm is supported.
//   5. Sets the output directory.
//   6. Ensures the output directory exists.
//   7. Decompresses the file and writes the decompressed files to the output directory.
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

// readAlgorithm reads the compression algorithm identifier from the provided
// compressed file reader. It first reads the length of the algorithm identifier
// and then reads the identifier itself.
//
// Parameters:
//   compressedFile (io.Reader): The reader from which the algorithm identifier
//   is to be read.
//
// Returns:
//   ([]byte, error): A byte slice containing the algorithm identifier if successful,
//   or an error if there was a problem reading from the file.
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

// setOutputDir sets the output directory to the directory of the first file if the output directory is not provided.
// 
// Parameters:
//   outputDir - A pointer to the string representing the output directory. If the string is empty, it will be set to the directory of the first file.
//   firstFilename - The name of the first file, used to determine the default output directory if outputDir is empty.
func setOutputDir(outputDir *string, firstFilename string) {
	// Set default output directory if not provided
	if *outputDir == "" {
		*outputDir = filepath.Dir(firstFilename)
	}
}

// walkDir traverses the directory specified by filenameStr and collects information
// about each file into the fileDataArr slice. It skips directories and only processes files.
// Each file's data is stored in a utils.FileData struct, which includes the file's name,
// size, and a reader for the file's contents.
//
// Parameters:
//   - filenameStr: The path of the directory to walk.
//   - fileDataArr: A pointer to a slice of utils.FileData where file information will be stored.
//
// Returns:
//   - error: An error if the directory walk fails or if there are issues opening files.
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