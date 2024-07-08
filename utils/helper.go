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
	BLUE   COLOR = "\033[1;34m%s\033[0m"
	WHITE  COLOR = "\033[1;37m%s\033[0m"
)

func ColorPrint(color COLOR, message string) {
	fmt.Printf(string(color), message)
}
