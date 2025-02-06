package lzma

import (
	"bytes"
	"encoding/binary"
	"errors"
	"file-compressor/constants"
	"file-compressor/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Zip(files []utils.FileData, output io.Writer) error {

	// Write the number of files
	if err := writeNumOfFiles(uint64(len(files)), output); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	for _, file := range files {
		reader := file.Reader

		//Compress and write the file name
		if err := writeFileName(file.Name, output); err != nil {
			return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}

		//write 64 bit 0 for the compressed size
		if err := binary.Write(output, binary.LittleEndian, uint64(0)); err != nil {
			return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}
		//Compress and write the data
		compressedLen, err := compressData(reader, output)

		if err != nil {
			return fmt.Errorf(constants.ERROR_COMPRESS, err)
		}

		//seek back to compressedLen bytes and write the compressed size
		if _, err := output.(io.Seeker).Seek(-int64(compressedLen+8), io.SeekCurrent); err != nil { // +4 for the 4 bytes of compressed size (uint64 -> 8 bytes) | 8bit = 1byte, 64bit = 8byte
			return fmt.Errorf("error seeking back to write the compressed size: %w", err)
		}

		if err := binary.Write(output, binary.LittleEndian, compressedLen); err != nil {
			return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
		}

		//seek back to the end of the file
		if _, err := output.(io.Seeker).Seek(0, io.SeekEnd); err != nil {
			return fmt.Errorf("error seeking to the end of the file: %w", err)
		}
	}

	return nil
}

func writeFileName(fileName string, output io.Writer) error {
	
	nameBuf := bytes.NewReader([]byte(fileName))

	compressedNameBuf := bytes.NewBuffer([]byte{})

	compLen, err := compressData(nameBuf, compressedNameBuf)
	if err != nil {
		return fmt.Errorf(constants.ERROR_COMPRESS, err)
	}

	// write length of the file name buffer
	if err := binary.Write(output, binary.LittleEndian, uint16(compLen)); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	// write the compressed file name
	if err := binary.Write(output, binary.LittleEndian, compressedNameBuf.Bytes()); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	return nil
}

func writeNumOfFiles(numOfFiles uint64, output io.Writer) error {

	if err := binary.Write(output, binary.LittleEndian, numOfFiles); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	return nil
}

func readNumOfFiles(input io.Reader) (uint64, error) {
	var numOfFiles uint64
	if err := binary.Read(input, binary.LittleEndian, &numOfFiles); err != nil {
		return 0, fmt.Errorf(constants.FILE_READ_ERROR, err)
	}

	return numOfFiles, nil
}

func Unzip(input io.Reader, outputPath string) ([]string, error) {

	if outputPath == "" {
		outputPath = "." // Use the current directory if no output path is provided
	}

	numOfFiles, err := readNumOfFiles(input)
	if err != nil {
		return nil, err
	}

	if numOfFiles < 1 {
		return nil, errors.New("no files to decompress")
	}

	filePaths := []string{}

	for i := uint64(0); i < numOfFiles; i++ {
		// get the file name
		fileName, err := readFileName(input)
		if err != nil {
			return nil, err
		}

		fileName = filepath.Join(outputPath, fileName)

		dir := filepath.Dir(fileName)

		if err := utils.MakeOutputDir(dir); err != nil {
			return nil, fmt.Errorf(constants.ERROR_CREATE_DIR, err)
		}

		// writer
		outputFile, err := os.Create(fileName)
		if err != nil {
			return nil, fmt.Errorf(constants.FILE_CREATE_ERROR, err)
		}

		// read the compressed size
		var compressedSize uint64
		if err := binary.Read(input, binary.LittleEndian, &compressedSize); err != nil {
			return nil, fmt.Errorf(constants.FILE_READ_ERROR, err)
		}

		err = decompressData(input, outputFile, compressedSize)
		if err != nil {
			return nil, fmt.Errorf(constants.ERROR_DECOMPRESS, err)
		}

		outputFile.Close()

		filePaths = append(filePaths, fileName)
	}

	return filePaths, nil
}

func readFileName(input io.Reader) (string, error) {

	var nameLen uint16
	if err := binary.Read(input, binary.LittleEndian, &nameLen); err != nil {
		return "", fmt.Errorf(constants.FILE_READ_ERROR, err)
	}

	buf := make([]byte, nameLen)
	if err := binary.Read(input, binary.LittleEndian, buf); err != nil {
		return "", fmt.Errorf(constants.FILE_READ_ERROR, err)
	}

	compressedFilename := bytes.NewBuffer(buf)

	nameBuffer := bytes.NewBuffer([]byte{})
	if err := decompressData(compressedFilename, nameBuffer, uint64(compressedFilename.Len())); err != nil {
		return "", fmt.Errorf(constants.ERROR_DECOMPRESS, err)
	}

	name := nameBuffer.String()

	return name, nil
}