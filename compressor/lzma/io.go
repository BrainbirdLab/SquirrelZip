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

// processFile handles the compression of a single file
func processFileCompression(file utils.FileData, output io.Writer, progressCallback utils.ProgressCallback) error {
	if progressCallback != nil {
		utils.UpdateProgress(utils.ProgressInfo{
			TotalFiles:      1,
			CurrentFile:     1,
			CurrentFileSize: file.Size,
			Message:         fmt.Sprintf("Compressing file: %s", filepath.Base(file.Name)),
		}, progressCallback)
	}

	if err := writeFileName(file.Name, output); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	if err := binary.Write(output, binary.LittleEndian, uint64(0)); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	compressedLen, err := compressDataWithProgress(file.Reader, output, filepath.Base(file.Name), progressCallback)
	if err != nil {
		return fmt.Errorf(constants.ERROR_COMPRESS, err)
	}

	if _, err := output.(io.Seeker).Seek(-int64(compressedLen+8), io.SeekCurrent); err != nil {
		return fmt.Errorf("error seeking back to write the compressed size: %w", err)
	}

	if err := binary.Write(output, binary.LittleEndian, compressedLen); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	if _, err := output.(io.Seeker).Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("error seeking to the end of the file: %w", err)
	}

	return nil
}

func Zip(files []utils.FileData, output io.Writer, progressCallback utils.ProgressCallback) error {
	if err := writeNumOfFiles(uint64(len(files)), output); err != nil {
		return fmt.Errorf(constants.FILE_WRITE_ERROR, err)
	}

	for i, file := range files {
		if err := processFileCompression(file, output, progressCallback); err != nil {
			return err
		}

		if progressCallback != nil {
			utils.UpdateProgress(utils.ProgressInfo{
				TotalFiles:      len(files),
				CurrentFile:     i + 1,
				CurrentFileSize: file.Size,
				ProcessedBytes:  file.Size,
				Message:         fmt.Sprintf("Completed: %s", filepath.Base(file.Name)),
			}, progressCallback)
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

// processFile handles the decompression of a single file
func processFile(input io.Reader, outputPath string, fileName string, progressCallback utils.ProgressCallback) (string, error) {
	fileName = filepath.Join(outputPath, fileName)
	dir := filepath.Dir(fileName)

	if err := utils.MakeOutputDir(dir); err != nil {
		return "", fmt.Errorf(constants.ERROR_CREATE_DIR, err)
	}

	outputFile, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf(constants.FILE_CREATE_ERROR, err)
	}
	defer outputFile.Close()

	var compressedSize uint64
	if err := binary.Read(input, binary.LittleEndian, &compressedSize); err != nil {
		return "", fmt.Errorf(constants.FILE_READ_ERROR, err)
	}

	if err := decompressDataWithProgress(input, outputFile, compressedSize, filepath.Base(fileName), progressCallback); err != nil {
		return "", fmt.Errorf(constants.ERROR_DECOMPRESS, err)
	}

	return fileName, nil
}

func Unzip(input io.Reader, outputPath string, progressCallback utils.ProgressCallback) ([]string, error) {
	if outputPath == "" {
		outputPath = "."
	}

	numOfFiles, err := readNumOfFiles(input)
	if err != nil {
		return nil, err
	}

	if numOfFiles < 1 {
		return nil, errors.New("no files to decompress")
	}

	filePaths := make([]string, 0, numOfFiles)

	for i := uint64(0); i < numOfFiles; i++ {
		fileName, err := readFileName(input)
		if err != nil {
			return nil, err
		}

		if progressCallback != nil {
			utils.UpdateProgress(utils.ProgressInfo{
				TotalFiles:  int(numOfFiles),
				CurrentFile: int(i + 1),
				Message:     fmt.Sprintf("Decompressing file: %s", filepath.Base(fileName)),
			}, progressCallback)
		}

		filePath, err := processFile(input, outputPath, fileName, progressCallback)
		if err != nil {
			return nil, err
		}

		filePaths = append(filePaths, filePath)

		if progressCallback != nil {
			utils.UpdateProgress(utils.ProgressInfo{
				TotalFiles:  int(numOfFiles),
				CurrentFile: int(i + 1),
				Message:     fmt.Sprintf("Completed: %s", filepath.Base(fileName)),
			}, progressCallback)
		}
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
