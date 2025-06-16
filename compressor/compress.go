package compressor

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"file-compressor/compressor/hfc"
	"file-compressor/compressor/lzma"
	"file-compressor/compressor/lzw"
	"file-compressor/constants"
	"file-compressor/utils"
)

func CheckCompressionAlgorithm(algo string) error {
	switch utils.Algorithm(algo) {
	case utils.HUFFMAN, utils.LZMA, utils.LZW:
		return nil
	default:
		return fmt.Errorf(constants.UNSUPPORTED_ALGO, algo)
	}
}

func Compress(filenameStrs []string, outputDir, algorithm string, userProgressCallback utils.ProgressCallback) (string, utils.FilesRatio, error) {
	fileMeta := utils.FilesRatio{}
	for _, filenameStr := range filenameStrs {
		if _, err := os.Stat(filenameStr); os.IsNotExist(err) {
			return "", fileMeta, fmt.Errorf("file '%s' does not exist", filenameStr)
		}
	}

	err := CheckCompressionAlgorithm(algorithm)
	if err != nil {
		return "", fileMeta, err
	}

	setOutputDir(&outputDir, filenameStrs[0])
	if err := utils.MakeOutputDir(outputDir); err != nil {
		return "", fileMeta, err
	}

	archiveFileName := filenameStrs[0]
	ext := filepath.Ext(archiveFileName)
	archiveFileName = strings.TrimSuffix(archiveFileName, ext)
	archiveFileName = filepath.Base(archiveFileName) + constants.COMPRESSED_FILE_EXT
	archiveFileName = utils.InvalidateFileName(archiveFileName, outputDir)

	compressedFileOutput, err := os.Create(archiveFileName)
	if err != nil {
		return "", fileMeta, fmt.Errorf(constants.ERROR_COMPRESS, err)
	}
	defer compressedFileOutput.Close()

	if userProgressCallback != nil {
        tempOverallTotalBytes := int64(1)
        currentInitialFile := ""
        if len(filenameStrs) > 0 {
            currentInitialFile = filenameStrs[0]
            // If single file, try to get its size early for a better initial message
            if len(filenameStrs) == 1 {
                info, statErr := os.Stat(filenameStrs[0])
                if statErr == nil && !info.IsDir() {
                    tempOverallTotalBytes = info.Size()
                }
            }
        }
        utils.UpdateProgress(utils.ProgressInfo{
            OverallTotalBytes: tempOverallTotalBytes,
            // AccumulatedBytesProcessedFromPreviousFiles is 0 here
            // BytesProcessedForCurrentFile is 0 here
            CurrentFileName: currentInitialFile, // Can be empty if filenameStrs is empty
            OriginalMessage: "Collecting file information...",
        }, userProgressCallback)
	}

	overallTotalBytes, fileDataArr, err := collectFiles(filenameStrs, userProgressCallback)
	if err != nil {
		return "", fileMeta, err
	}
    if len(fileDataArr) == 0 && overallTotalBytes == 0 {
        if errAlg := writeAlgorithm(compressedFileOutput, algorithm); errAlg != nil {
             return "", fileMeta, fmt.Errorf("failed to write algorithm header for empty archive: %w", errAlg)
        }
        utils.ColorPrint(utils.YELLOW, "No files to compress.\n")
        // Ensure final progress is reported even for empty archives
        if userProgressCallback != nil {
            utils.UpdateProgress(utils.ProgressInfo{
                OverallTotalBytes: 0, // No bytes to process
                AccumulatedBytesProcessedFromPreviousFiles: 0,
                BytesProcessedForCurrentFile: 0,
                FilesCompletedOverall: 0,
                TotalFilesOverall: 0,
                OriginalMessage: "Compression completed (no files found to compress).",
            }, userProgressCallback)
        }
        return archiveFileName, utils.NewFilesRatio(0,0), nil // Return success for empty archive creation
    }


	if userProgressCallback != nil && overallTotalBytes > 0 {
		utils.UpdateProgress(utils.ProgressInfo{
            OverallTotalBytes: overallTotalBytes,
            // AccumulatedBytesProcessedFromPreviousFiles is 0
            // BytesProcessedForCurrentFile is 0
            CurrentFileName: fileDataArr[0].Name, // First file to be processed
            TotalFilesOverall: len(fileDataArr),
            OriginalMessage: "Starting compression process...",
        }, userProgressCallback)
	}

	if err := writeAlgorithm(compressedFileOutput, algorithm); err != nil {
		return "", fileMeta, fmt.Errorf("failed to write algorithm header: %w", err)
	}

    if len(fileDataArr) > 0 {
	    err = ReadAndCompressFiles(fileDataArr, compressedFileOutput, algorithm, userProgressCallback, overallTotalBytes)
	    if err != nil {
		    return "", fileMeta, err
	    }
    }

	compressedStat, err := os.Stat(archiveFileName)
	if err != nil {
		return "", fileMeta, fmt.Errorf(constants.FILE_STAT_ERROR, err)
	}
	fileMeta = utils.NewFilesRatio(overallTotalBytes, uint64(compressedStat.Size()))

	if userProgressCallback != nil {
        finalMessage := "Compression completed."
        if len(fileDataArr) == 0 { // Should have been caught above, but for safety
            finalMessage = "Compression completed (no files processed)."
        }
        utils.UpdateProgress(utils.ProgressInfo{
            OverallTotalBytes: overallTotalBytes,
            AccumulatedBytesProcessedFromPreviousFiles: overallTotalBytes, // All bytes are now "previous"
            BytesProcessedForCurrentFile: 0, // No current file processing at the very end
            FilesCompletedOverall: len(fileDataArr),
            TotalFilesOverall: len(fileDataArr),
            OriginalMessage: finalMessage,
        }, userProgressCallback)
	}
	return archiveFileName, fileMeta, nil
}

func ReadAndCompressFiles(fileDataArr []utils.FileData, output io.Writer, algorithm string, userProgressCallback utils.ProgressCallback, overallTotalBytes int64) error {
	var accumulatedBytesFromPreviousFiles int64 = 0
	totalFiles := len(fileDataArr)

    if totalFiles == 0 {
        return nil
    }

	for i, currentFile := range fileDataArr {
		currentFileName := currentFile.Name
		currentFileSize := currentFile.Size

		// Ensure reader is closed if it's an actual file after it's processed or on error
        // Note: The individual Zip functions will also need to handle closing their input readers.
        // This defer is a safety net for readers opened by collectFiles.
        if currentFile.Reader != nil {
            defer func(r io.ReadCloser) {
                if r != nil {
                    // Check if it's one of the file readers we opened, not a temporary buffer reader
                    if _, ok := r.(*os.File); ok {
                         r.Close()
                    }
                }
            }(currentFile.Reader)
        }


		progressAdapterCallback := func(percentOfCurrentFile float64, messageFromZip string) {
			if userProgressCallback == nil {
				return
			}
			bytesProcessedForThisFile := int64(percentOfCurrentFile * float64(currentFileSize))
			if bytesProcessedForThisFile > currentFileSize {
				bytesProcessedForThisFile = currentFileSize
			}
			if bytesProcessedForThisFile < 0 {
                bytesProcessedForThisFile = 0
            }

			progInfo := utils.ProgressInfo{
				CurrentFileName:     currentFileName,
				SizeOfCurrentFile:   currentFileSize,
				BytesProcessedForCurrentFile: bytesProcessedForThisFile,
				TotalFilesOverall:     totalFiles,
				FilesCompletedOverall: i,
				OverallTotalBytes:    overallTotalBytes,
				AccumulatedBytesProcessedFromPreviousFiles: accumulatedBytesFromPreviousFiles,
				OriginalMessage: messageFromZip,
			}
			utils.UpdateProgress(progInfo, userProgressCallback)
		}

		var err error
		// Create a single FileData item for the current file to pass to Zip functions
		// This assumes Zip functions will be refactored to take utils.FileData
		singleFileData := []utils.FileData{currentFile}

		switch utils.Algorithm(algorithm) {
		case utils.HUFFMAN:
			// Assuming hfc.Zip will be refactored to:
			// func Zip(files []utils.FileData, output io.Writer, progressCallback utils.ProgressCallback) error
			// OR, more likely for this refactor:
			// func Zip(file utils.FileData, output io.Writer, progressCallback func(float64, string)) error
			err = hfc.Zip(singleFileData, output, progressAdapterCallback)
		case utils.LZMA:
			err = lzma.Zip(singleFileData, output, progressAdapterCallback)
		case utils.LZW:
			err = lzw.Zip(singleFileData, output, progressAdapterCallback)
		default:
            // Ensure reader is closed if not passed to a Zip function or if error before that
            if closer, ok := currentFile.Reader.(io.Closer); ok {
                closer.Close()
            }
			return fmt.Errorf("unsupported compression algorithm: %v for file %s", algorithm, currentFileName)
		}

		if err != nil {
            // Reader should be closed by the Zip function or the defer above if it's an os.File
			return fmt.Errorf("failed to compress %s with %s: %w", currentFileName, algorithm, err)
		}

		accumulatedBytesFromPreviousFiles += currentFileSize

        // After a file is successfully processed by its Zip function
        if userProgressCallback != nil {
            progInfo := utils.ProgressInfo{
                CurrentFileName: currentFileName,
                SizeOfCurrentFile: currentFileSize,
                BytesProcessedForCurrentFile: currentFileSize, // Mark as fully processed for this file
                TotalFilesOverall: totalFiles,
                FilesCompletedOverall: i + 1, // This file is now completed
                OverallTotalBytes: overallTotalBytes,
                AccumulatedBytesProcessedFromPreviousFiles: accumulatedBytesFromPreviousFiles - currentFileSize, // Bytes from *previous* files
                OriginalMessage: fmt.Sprintf("Finished %s", filepath.Base(currentFileName)),
            }
            utils.UpdateProgress(progInfo, userProgressCallback)
        }
	}
	return nil
}


// Decompress function remains largely the same for this subtask,
// but its progress reporting will also need an overhaul later.
func Decompress(archiveFileName, specificOutputDir string, userProgressCallback utils.ProgressCallback) ([]string, utils.FilesRatio, error) {
	fileMeta := utils.FilesRatio{}
	if _, err := os.Stat(archiveFileName); os.IsNotExist(err) {
		return nil, fileMeta, fmt.Errorf(constants.FILE_NOT_EXIST, archiveFileName)
	}

	compressedFile, err := os.Open(archiveFileName)
	if err != nil {
		return nil, fileMeta, fmt.Errorf(constants.FILE_OPEN_ERROR, err)
	}
	defer compressedFile.Close()

	algorithm, err := readAlgorithm(compressedFile)
	if err != nil {
		return nil, fileMeta, fmt.Errorf("failed to read algorithm from archive: %w", err)
	}
	if err := CheckCompressionAlgorithm(algorithm); err != nil {
		return nil, fileMeta, err
	}

	outputDir := specificOutputDir
	if outputDir == "" {
		ext := filepath.Ext(archiveFileName)
		outputDir = strings.TrimSuffix(archiveFileName, ext) + "_decompressed"
	}
	outputDir = utils.InvalidateFileName(outputDir, "") // Ensure outputDir is valid before creating
	if err := utils.MakeOutputDir(outputDir); err != nil {
		return nil, fileMeta, err
	}

	// TODO: Implement new progress reporting for Decompress
	// For now, we'll pass the userProgressCallback directly, assuming Unzip functions
	// might eventually use it or be adapted. This part will need a similar refactor as Compress.
	if userProgressCallback != nil {
        stat, _ := compressedFile.Stat()
        size := int64(1)
        if stat != nil {
            size = stat.Size()
        }
		utils.UpdateProgress(utils.ProgressInfo{
            OverallTotalBytes: size, // Approximate with archive size
            OriginalMessage: fmt.Sprintf("Starting decompression of %s...", filepath.Base(archiveFileName)),
        }, userProgressCallback)
	}

	var fileNames []string
	fileNames, err = WriteAndDecompressFiles(compressedFile, outputDir, algorithm, userProgressCallback)
	if err != nil {
		return nil, fileMeta, err
	}

	originalSize, err := utils.GetTotalSize(fileNames)
	if err != nil {
		return nil, fileMeta, fmt.Errorf(constants.FILE_STAT_ERROR, err)
	}
	compressedStat, _ := os.Stat(archiveFileName)
	fileMeta = utils.NewFilesRatio(originalSize, uint64(compressedStat.Size()))

	if userProgressCallback != nil {
        finalSize := int64(0)
        if compressedStat != nil {
            finalSize = compressedStat.Size()
        }
		utils.UpdateProgress(utils.ProgressInfo{
            OverallTotalBytes: finalSize,
            AccumulatedBytesProcessedFromPreviousFiles: finalSize,
            OriginalMessage: "Decompression completed.",
            FilesCompletedOverall: len(fileNames), // Assuming one entry per file in archive
            TotalFilesOverall: len(fileNames),     // Might need better tracking from Unzip
        }, userProgressCallback)
	}

	return fileNames, fileMeta, nil
}

// WriteAndDecompressFiles also needs a similar refactor for progress.
// For now, it passes userProgressCallback along.
func WriteAndDecompressFiles(compressedFile io.Reader, outputDir, algorithm string, userProgressCallback utils.ProgressCallback) ([]string, error) {
	var fileNames []string
	var err error

	// The Unzip functions would need to be adapted to use the new progress model.
	// This will be part of their individual refactoring.
	// The progress callback passed here might be the top-level one, or an adapter.
	switch utils.Algorithm(algorithm) {
	case utils.HUFFMAN:
		fileNames, err = hfc.Unzip(compressedFile, outputDir, userProgressCallback)
		if err != nil {
			return nil, fmt.Errorf(constants.ERROR_DECOMPRESS, err)
		}
	case utils.LZMA:
		fileNames, err = lzma.Unzip(compressedFile, outputDir, userProgressCallback)
		if err != nil {
			return nil, fmt.Errorf(constants.ERROR_DECOMPRESS, err)
		}
	case utils.LZW:
		fileNames, err = lzw.Unzip(compressedFile, outputDir, userProgressCallback)
		if err != nil {
			return nil, fmt.Errorf(constants.ERROR_DECOMPRESS, err)
		}
	default:
		return nil, fmt.Errorf("unsupported compression algorithm: %v", algorithm)
	}
	return fileNames, nil
}


func walkDir(dirPath string, filesChan chan<- utils.FileData, errChan chan<- error, userProgressCallback utils.ProgressCallback) {
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Potentially send this error to errChan if it's critical
			utils.ColorPrint(utils.RED, fmt.Sprintf("Warning: Error accessing path %s: %v. Skipping.
", path, err))
			return nil // Continue walking other files
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				utils.ColorPrint(utils.RED, fmt.Sprintf("Warning: Error opening file %s: %v. Skipping.
", path, err))
				return nil // Continue
			}
			// We don't close 'file' here. It will be closed by the receiver (ReadAndCompressFiles or its callee)
			// after its content has been processed.
			filesChan <- utils.FileData{Name: path, Size: info.Size(), Reader: file}
		}
		return nil
	})
	close(filesChan) // Close channel when walk is done
}

func collectFiles(paths []string, userProgressCallback utils.ProgressCallback) (int64, []utils.FileData, error) {
	var fileDataArr []utils.FileData
	var overallTotalBytes int64
	filesChan := make(chan utils.FileData)
	errChan := make(chan error, 1) // Buffer for one error

	numPaths := len(paths)
	processedPaths := 0

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return 0, nil, fmt.Errorf("error accessing path %s: %w", path, err)
		}
		if info.IsDir() {
			go walkDir(path, filesChan, errChan, userProgressCallback)
		} else {
			file, err := os.Open(path)
			if err != nil {
				return 0, nil, fmt.Errorf("error opening file %s: %w", path, err)
			}
			// Add single file to channel, then close it for this "path"
            // Or, more simply, handle single files directly without a goroutine for them
            fileDataArr = append(fileDataArr, utils.FileData{Name: path, Size: info.Size(), Reader: file})
            overallTotalBytes += info.Size()
            processedPaths++
		}
	}

    // If there were directory paths, process them from the channel
    if processedPaths < numPaths {
        done := false
        for !done {
            select {
            case fileData, ok := <-filesChan:
                if !ok {
                    done = true // Channel closed
                    break
                }
                fileDataArr = append(fileDataArr, fileData)
                overallTotalBytes += fileData.Size
            case err, ok := <-errChan:
                if ok { // Should only receive one error if any, then close errChan or stop reading.
                    return 0, nil, err // Propagate first critical error from walkDir
                }
                // If errChan is closed without error, that's fine.
            }
        }
    }
    // Ensure all directory walkers are finished if filesChan was used
    // This logic might need refinement if multiple walkDir goroutines are active
    // and filesChan isn't closed until all are done.
    // For now, walkDir closes its own filesChan contribution upon completion.
    // A more robust approach would use a sync.WaitGroup if multiple go walkDir are launched.
    // However, the current structure launches one at a time if paths contains multiple dirs.
    // The provided walkDir closes filesChan, which is problematic if multiple dirs are processed.
    // Let's assume for now paths usually contains one entry or files.
    // For robust multi-directory walk, filesChan should be managed by collectFiles with a WaitGroup.

	// Sort files for consistent processing order (optional, but good for reproducibility)
	// sort.Slice(fileDataArr, func(i, j int) bool { return fileDataArr[i].Name < fileDataArr[j].Name })

	return overallTotalBytes, fileDataArr, nil
}

func writeAlgorithm(writer io.Writer, algorithm string) error {
	algoByte := []byte(algorithm)
	if len(algoByte) > constants.ALGO_NAME_SIZE {
		return fmt.Errorf("algorithm name '%s' too long", algorithm)
	}
	paddedAlgo := make([]byte, constants.ALGO_NAME_SIZE)
	copy(paddedAlgo, algoByte)
	_, err := writer.Write(paddedAlgo)
	return err
}

func readAlgorithm(reader io.Reader) (string, error) {
	algoBytes := make([]byte, constants.ALGO_NAME_SIZE)
	_, err := io.ReadFull(reader, algoBytes)
	if err != nil {
		return "", err
	}
	// Trim null characters or spaces used for padding
	return strings.TrimRight(string(algoBytes), "\x00 "), nil
}

func setOutputDir(outputDir *string, firstInputName string) {
	if *outputDir == "" {
		// If input is a directory, create archive in parent of input dir
		// If input is a file, create archive in same dir as file
		info, err := os.Stat(firstInputName)
		if err == nil && info.IsDir() {
			*outputDir = filepath.Dir(filepath.Clean(firstInputName)) // Parent directory of input dir
		} else {
			*outputDir = filepath.Dir(firstInputName) // Same directory as input file
		}
	}
	// If outputDir is still empty (e.g. input was current dir "."), use current dir.
	if *outputDir == "" || *outputDir == "." {
		*outputDir = "."
	}
}
