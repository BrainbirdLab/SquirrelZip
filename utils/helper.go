package utils

import (
	"file-compressor/constants"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileData struct {
	Name   string
	Size   int64
	Reader io.Reader
}

type Algorithm string

const (
	HUFFMAN Algorithm = "huffman"
	ARITHMETIC Algorithm = "arithmetic"

	UNSUPPORTED Algorithm = "unsupported"
)

const (
	FailedToCompress string = "failed to compress data: %v"
	FailedToDecompress string = "failed to decompress data: %v"
	MissMatch string = "decompressed data does not match original data: %v != %v"
)

type COLOR string

const (
	GREY  COLOR = "\033[1;30m%s\033[0m"
	RED    COLOR = "\033[1;31m%s\033[0m"
	GREEN  COLOR = "\033[1;32m%s\033[0m"
	YELLOW COLOR = "\033[1;33m%s\033[0m"
	PURPLE COLOR = "\033[1;35m%s\033[0m"
	CYAN   COLOR = "\033[1;36m%s\033[0m"
	BLUE   COLOR = "\033[1;34m%s\033[0m"
	WHITE  COLOR = "\033[1;37m%s\033[0m"
)

func ColorPrint(color COLOR, message string) {
	fmt.Printf(string(color), message)
}

func MakeOutputDir(outputDir string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.Mkdir(outputDir, 0777)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}
	return nil
}

func FileSize(sizeBytes uint64) string {
	var unit uint64 = 1024
	if sizeBytes < unit {
		return fmt.Sprintf("%d B", sizeBytes)
	}
	div, exp := uint64(unit), 0
	for n := sizeBytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(sizeBytes)/float64(div), "KMGTPE"[exp])
}

func TimeTrack(startTime, endTime time.Time) string {
	elapsedTime := endTime.Sub(startTime)
	//return nanoseconds, microseconds, milliseconds, seconds, minutes, hours
	if elapsedTime < 1000 {
		return fmt.Sprintf("%d ns", elapsedTime)
	}

	if elapsedTime < 1000000 {
		return fmt.Sprintf("%d µs", elapsedTime/1000)
	}

	if elapsedTime < 1000000000 {
		return fmt.Sprintf("%d ms", elapsedTime/1000000)
	}

	if elapsedTime < 60000000000 {
		return fmt.Sprintf("%d s", elapsedTime/1000000000)
	}

	if elapsedTime < 3600000000000 {
		return fmt.Sprintf("%d m", elapsedTime/60000000000)
	}

	return fmt.Sprintf("%d h", elapsedTime/3600000000000)
}

type FilesRatio struct {
	inital uint64
	compressed uint64
}

func NewFilesRatio(initial, compressed uint64) FilesRatio {
	return FilesRatio{
		inital: initial,
		compressed: compressed,
	}
}

func (f *FilesRatio) PrintFileInfo() {
	fmt.Printf("Target size: %s\n", FileSize(f.inital))
	fmt.Printf("Compressed size: %s\n", FileSize(f.compressed))
}

func (f *FilesRatio) PrintCompressionRatio() {
	compressionRatio := (float64(f.compressed) / float64(f.inital))  * 100
	fmt.Printf("Compression ratio: %.2f%%\n", compressionRatio)
}

func InvalidateFileName(filename string, outputDir string) string {
	fileExt := filepath.Ext(filename)
	//extract the file name without the extension
	filename = filepath.Base(filename)
	originalName := strings.TrimSuffix(filename, fileExt)

	finalFile := filepath.Join(outputDir, originalName + fileExt)

	count := 1
	for {
		//if file already exists, add a number to the filename before the extension and check again
		if _, err := os.Stat(finalFile); err == nil {
			filename = fmt.Sprintf("%s_%d", originalName, count)
			finalFile = filepath.Join(outputDir, filename + fileExt)
		} else {
			break
		}
		count++
	}
	return finalFile
}

func SafeDeleteFile(filePath string) {
	err := os.Remove(filePath)
	if err != nil {
		ColorPrint(RED, fmt.Sprintf(constants.FILE_REMOVE_ERROR, err.Error()) + "\n")
	}
}