package utils

import (
	"fmt"
	"time"
)

type File struct {
	Name    string
	Content []byte
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

func FileSize(sizeBytes int64) string {
	var unit int64 = 1024
	if sizeBytes < unit {
		return fmt.Sprintf("%d B", sizeBytes)
	}
	div, exp := int64(unit), 0
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
	inital int
	compressed int
}

func NewFilesRatio(initial, compressed int) FilesRatio {
	return FilesRatio{
		inital: 0,
		compressed: 0,
	}
}

func (f FilesRatio) PrintFileInfo() {
	fmt.Printf("Target size: %d\n", f.inital)
	fmt.Printf("Compressed size: %d\n", f.compressed)
}

func (f FilesRatio) PrintCompressionRatio() {
	compressionRatio := float64(f.inital) / float64(f.compressed) * 100
	fmt.Printf("Compression ratio: %.2f%%\n", compressionRatio)
}