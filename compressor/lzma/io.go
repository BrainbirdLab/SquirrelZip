package lzma

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"file-compressor/constants"
	"file-compressor/utils"
)

func Zip(files []utils.FileData, output io.Writer) error {
	// Write number of files
	if err := binary.Write(output, binary.LittleEndian, uint64(len(files))); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	for _, file := range files {
		// Write filename length and filename
		filename := []byte(file.Name)
		if err := binary.Write(output, binary.LittleEndian, uint16(len(filename))); err != nil {
			return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}
		if _, err := output.Write(filename); err != nil {
			return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}

		// Read file data
		data, err := io.ReadAll(file.Reader)
		if err != nil {
			return fmt.Errorf(constants.FILE_READ_ERROR, err)
		}

		// Compress data
		compressed, err := compressData(data)
		if err != nil {
			return fmt.Errorf(constants.ERROR_COMPRESS, err)
		}

		// Write compressed data length and data
		if err := binary.Write(output, binary.LittleEndian, uint64(len(compressed))); err != nil {
			return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}
		if _, err := output.Write(compressed); err != nil {
			return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}
	}

	return nil
}

func Unzip(input io.Reader, outputPath string) ([]string, error) {
	if outputPath == "" {
		outputPath = "."
	}

	// Read number of files
	var numFiles uint64
	if err := binary.Read(input, binary.LittleEndian, &numFiles); err != nil {
		return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
	}

	if numFiles < 1 {
		return nil, errors.New("no files to decompress")
	}

	filePaths := make([]string, 0, numFiles)

	for i := uint64(0); i < numFiles; i++ {
		// Read filename length and filename
		var nameLen uint16
		if err := binary.Read(input, binary.LittleEndian, &nameLen); err != nil {
			return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
		}

		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(input, nameBytes); err != nil {
			return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
		}

		fileName := filepath.Join(outputPath, string(nameBytes))
		dir := filepath.Dir(fileName)

		// Create output directory if needed
		if err := utils.MakeOutputDir(dir); err != nil {
			return nil, fmt.Errorf(constants.ERROR_CREATE_DIR, err)
		}

		// Read compressed data length and data
		var compressedLen uint64
		if err := binary.Read(input, binary.LittleEndian, &compressedLen); err != nil {
			return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
		}

		compressedData := make([]byte, compressedLen)
		if _, err := io.ReadFull(input, compressedData); err != nil {
			return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
		}

		// Decompress data
		decompressed, err := decompressData(compressedData)
		if err != nil {
			return nil, fmt.Errorf(constants.ERROR_DECOMPRESS, err)
		}

		// Write decompressed data to file
		outputFile, err := os.Create(fileName)
		if err != nil {
			return nil, fmt.Errorf(constants.FILE_CREATE_ERROR, err)
		}

		if _, err := io.Copy(outputFile, bytes.NewReader(decompressed)); err != nil {
			outputFile.Close()
			return nil, fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}
		outputFile.Close()

		filePaths = append(filePaths, fileName)
	}

	return filePaths, nil
}
