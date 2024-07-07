package utils

import (
    "flag"
    "fmt"
    "os"
)

type Options struct {
    InputFile  string
    OutputFile string
    Password   string
}

func ParseCommandLine() *Options {
    inputFile := flag.String("i", "", "Input file to be compressed")
    outputFile := flag.String("o", "compressed.pcd", "Output compressed file name")
    password := flag.String("p", "", "Password for encryption (optional)")
    flag.Parse()

    if *inputFile == "" {
        fmt.Println("Input file is required.")
        flag.Usage()
        os.Exit(1)
    }

    return &Options{
        InputFile:  *inputFile,
        OutputFile: *outputFile,
        Password:   *password,
    }
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