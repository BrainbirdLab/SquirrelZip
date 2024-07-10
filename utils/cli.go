//parseCLI.go

package utils

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
)

func ParseCLI() ([]string, *string, *string, *bool) {
	// CLI arguments
	version := flag.Bool("v", false, "Print version")
	inputFiles := flag.String("i", "", "Input files or directory to be compressed")
	outputDir := flag.String("o", "", "Output directory for compressed files (Optional)")
	password := flag.String("p", "", "Password for encryption (Optional)")
	readAllFiles := flag.Bool("a", false, "Read all files in the input directory")
	decompressMode := flag.Bool("d", false, "Decompress mode")
	flag.Parse()

	if *version {
		ColorPrint(WHITE, "---------- PI ARCHIVER ----------\n")
		ColorPrint(YELLOW, "Version: v1.0.0\n")
		// dev info
		ColorPrint(WHITE, "Developed by: https://github.com/itsfuad/\n")
		ColorPrint(WHITE, "---------------------------------")
		os.Exit(0)
	}

	if *inputFiles == "" {
		ColorPrint(RED, "No input files or directory provided\n")
		flag.Usage()
		os.Exit(1)
	}

	var filenameStrs []string

	// Handle reading all files in the input directory
	if *readAllFiles {
		// Check if input is a directory
		info, err := os.Stat(*inputFiles)
		if os.IsNotExist(err) {
			ColorPrint(RED, "Input directory does not exist.\n")
			os.Exit(1)
		}

		if !info.IsDir() {
			ColorPrint(RED, "Provided input is not a directory.\n")
			os.Exit(1)
		}

		// Read all files in the directory
		err = filepath.Walk(*inputFiles, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				filenameStrs = append(filenameStrs, path)
			}
			return nil
		})
		if err != nil {
			ColorPrint(RED, err.Error())
			os.Exit(1)
		}
	} else {
		// Split input files by comma and trim spaces and quotes (if any)
		for _, filename := range strings.Split(*inputFiles, ",") {
			filenameStrs = append(filenameStrs, strings.Trim(filename, " '\""))
		}
	}

	return filenameStrs, outputDir, password, decompressMode
}
