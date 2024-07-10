package utils

import (
	"fmt"
)

type File struct {
	Name    string
	Content []byte
}

type COLOR string

const (
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
	return fmt.Sprintf("%.1f %ciB",
		float64(sizeBytes)/float64(div), "KMGTPE"[exp])
}
