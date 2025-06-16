package lzw

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"file-compressor/constants"
	"file-compressor/utils"
)

// Zip compresses a single file using LZW algorithm and writes to the output writer.
// It writes metadata for the file (name length, name, data length, data).
// The progressCallback expects (percentOfCurrentFile float64, messageFromZip string).
func Zip(file utils.FileData, output io.Writer, progressCallback func(percentOfCurrentFile float64, messageFromZip string)) error {
	if progressCallback != nil {
		// Initial message for this specific file
		progressCallback(0.0, fmt.Sprintf("Starting LZW for %s", filepath.Base(file.Name)))
	}

	if file.Reader == nil {
		if file.Size > 0 {
			utils.ColorPrint(utils.YELLOW, fmt.Sprintf("Warning: LZW Zip skipping '%s' due to nil reader.\n", file.Name))
		}
		// If reader is nil, we can't proceed.
        // Report 100% for this "skipped" file to allow overall progress to advance.
        if progressCallback != nil {
            progressCallback(1.0, fmt.Sprintf("Skipped %s (nil reader)", filepath.Base(file.Name)))
        }
		return nil
	}

	fileReader := file.Reader

	defer func(r io.Reader) {
		if c, ok := r.(io.Closer); ok {
			c.Close()
		}
	}(fileReader)

	fileNameBytes := []byte(filepath.Base(file.Name))
	if err := binary.Write(output, binary.LittleEndian, uint32(len(fileNameBytes))); err != nil {
		return fmt.Errorf("LZW Zip: failed to write file name length for %s: %w", file.Name, err)
	}

	if _, err := output.Write(fileNameBytes); err != nil {
		return fmt.Errorf("LZW Zip: failed to write file name for %s: %w", file.Name, err)
	}

	tempCompressedData := utils.NewResizableBuffer()

    var bytesReadForThisFile int64
    onProgressUpdate := func(readBytes int64, totalBytes int64) {
        bytesReadForThisFile = readBytes
        if progressCallback != nil && totalBytes > 0 {
            percent := float64(readBytes) / float64(totalBytes)
            if percent > 1.0 { percent = 1.0 } // Cap percent at 1.0
            progressCallback(percent, fmt.Sprintf("Processing %s", filepath.Base(file.Name)))
        }
    }

    trackingReader := utils.NewProgressReader(fileReader, file.Size, onProgressUpdate)

	errCompress := compressData(trackingReader, tempCompressedData)
	if errCompress != nil {
		return fmt.Errorf("LZW Zip: failed to compress data for %s: %w", file.Name, errCompress)
	}

	// After compression, all bytes of this file should have been read (or an error occurred).
    // Ensure a final 100% progress for this file is signaled if no error.
    if progressCallback != nil && file.Size > 0 {
        // Check if the last report was already 100%
        // This might be redundant if compressData + ProgressTrackingReader already guarantees a final 100% call,
        // but it's a safeguard.
        currentProgress := 0.0
        if file.Size > 0 { // Avoid division by zero if file.Size is 0
            currentProgress = float64(bytesReadForThisFile) / float64(file.Size)
        }
        if currentProgress < 1.0 {
            progressCallback(1.0, fmt.Sprintf("Finalizing %s", filepath.Base(file.Name)))
        }
    }


	compressedSize := uint64(tempCompressedData.Len())
	if err := binary.Write(output, binary.LittleEndian, compressedSize); err != nil {
		return fmt.Errorf("LZW Zip: failed to write compressed data length for %s: %w", file.Name, err)
	}

	if _, err := io.Copy(output, tempCompressedData.Reader()); err != nil {
		return fmt.Errorf("LZW Zip: failed to write compressed data for %s: %w", file.Name, err)
	}

	if progressCallback != nil {
		// Final confirmation for this specific file being done.
		progressCallback(1.0, fmt.Sprintf("Finished LZW for %s", filepath.Base(file.Name)))
	}
	return nil
}

// Unzip decompresses data using LZW algorithm from the input reader and writes to outputDir.
func Unzip(input io.Reader, outputDir string, progressCallback utils.ProgressCallback) ([]string, error) {
	if progressCallback != nil {
		// This is a general progress callback, not the fine-grained one.
        // We'd need to adapt Unzip similarly to provide detailed progress.
        // For now, just initial and final messages based on the old system.
		// Example of how it might be adapted later:
        // utils.UpdateProgress(utils.ProgressInfo{OriginalMessage: "Starting LZW decompression..."}, progressCallback)
		progressCallback(0.0, "Starting LZW decompression...")
	}
	var decompressedFiles []string
	fileIndex := 0

	for {
		fileIndex++
		var fileNameLen uint32
		err := binary.Read(input, binary.LittleEndian, &fileNameLen)
		if err != nil {
			if err == io.EOF {
				break
			}
			return decompressedFiles, fmt.Errorf("LZW Unzip: failed to read file name length (file #%d): %w", fileIndex, err)
		}

		if fileNameLen == 0 || fileNameLen > 1024*1024 {
             return decompressedFiles, fmt.Errorf("LZW Unzip: invalid file name length %d (file #%d), archive might be corrupted", fileNameLen, fileIndex)
        }

		fileNameBytes := make([]byte, fileNameLen)
		if _, err := io.ReadFull(input, fileNameBytes); err != nil {
			return decompressedFiles, fmt.Errorf("LZW Unzip: failed to read file name (file #%d): %w", fileIndex, err)
		}
		fileName := string(fileNameBytes)
        cleanFileName := filepath.Clean(fileName)

        if strings.HasPrefix(cleanFileName, ".."+string(filepath.Separator)) || cleanFileName == ".." || strings.HasPrefix(cleanFileName, string(filepath.Separator)) || strings.Contains(fileName, "\\") || strings.Contains(fileName, ":") {
            utils.ColorPrint(utils.RED, fmt.Sprintf("Warning: LZW Unzip potentially unsafe file path '%s' in archive. Skipping.\n", fileName))
            var compressedSizeToSkip uint64
            if errSize := binary.Read(input, binary.LittleEndian, &compressedSizeToSkip); errSize != nil {
                 return decompressedFiles, fmt.Errorf("LZW Unzip: failed to read size for potentially unsafe file %s: %w", fileName, errSize)
            }
            if _, errSeek := io.CopyN(io.Discard, input, int64(compressedSizeToSkip)); errSeek != nil {
                 return decompressedFiles, fmt.Errorf("LZW Unzip: failed to skip data for potentially unsafe file %s: %w", fileName, errSeek)
            }
            // TODO: Report progress for skipped file if possible, e.g.
            // if progressCallback != nil { utils.UpdateProgress(utils.ProgressInfo{ ... CurrentFileName: fileName, OriginalMessage: "Skipped unsafe file" ... }, progressCallback) }
            continue
        }
		filePath := filepath.Join(outputDir, cleanFileName)

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return decompressedFiles, fmt.Errorf("LZW Unzip: failed to create directory for %s: %w", filePath, err)
		}

		var compressedSize uint64
		if err := binary.Read(input, binary.LittleEndian, &compressedSize); err != nil {
			return decompressedFiles, fmt.Errorf("LZW Unzip: failed to read compressed data length for %s: %w", fileName, err)
		}

		outFile, createErr := os.Create(filePath)
		if createErr != nil {
			return decompressedFiles, fmt.Errorf("LZW Unzip: %s for %s: %w", constants.FILE_CREATE_ERROR, filePath, createErr)
		}

		successfulDecompression := false
		defer func(file *os.File, path string) {
			file.Close()
			if !successfulDecompression {
				os.Remove(path)
			}
		}(outFile, filePath)

		if compressedSize == 0 {
            successfulDecompression = true
            decompressedFiles = append(decompressedFiles, filePath)
            if progressCallback != nil {
				// This is an approximation of overall progress, not ideal for the new system
				progress := float64(fileIndex) / (float64(fileIndex) + 1.0)
                // Example: utils.UpdateProgress(utils.ProgressInfo{ ... CurrentFileName: fileName, OriginalMessage: "Decompressed empty file" ... }, progressCallback)
				progressCallback(progress, fmt.Sprintf("Decompressed %s (empty)", fileName))
            }
            continue
        }

		limitedReader := io.LimitReader(input, int64(compressedSize))
        // TODO: Wrap limitedReader with utils.NewProgressReader for decompressData if we want byte-level progress for decompression
		errDecompress := decompressData(limitedReader, outFile)

		if errDecompress != nil {
			return decompressedFiles, fmt.Errorf("LZW Unzip: failed to decompress data for %s: %w", fileName, errDecompress)
		}
		successfulDecompression = true

		decompressedFiles = append(decompressedFiles, filePath)
		if progressCallback != nil {
			// Approximation
			progress := float64(fileIndex) / (float64(fileIndex) + 1.0)
            // Example: utils.UpdateProgress(utils.ProgressInfo{ ... CurrentFileName: fileName, OriginalMessage: "Decompressed file" ... }, progressCallback)
			progressCallback(progress, fmt.Sprintf("Decompressed %s", fileName))
		}
	}

	if progressCallback != nil {
		// Example: utils.UpdateProgress(utils.ProgressInfo{OriginalMessage: "LZW decompression completed.", OverallProcessedBytes: totalArchiveSize, OverallTotalBytes: totalArchiveSize}, progressCallback)
		progressCallback(1.0, "LZW decompression completed.")
	}
	return decompressedFiles, nil
}
