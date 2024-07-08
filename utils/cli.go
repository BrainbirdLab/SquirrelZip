package utils

import (
	"flag"
	"os"
	"strings"
)


func ParseCLI() ([]string, *string, *string, *bool) {
	//cli arguments
	inputFiles := flag.String("i", "", "Input files to be compressed")
	outputDir := flag.String("o", "", "Output directory for compressed files (Optional)")
	password := flag.String("p", "", "Password for encryption (Optional)")
	readAllFiles := flag.Bool("a", false, "Read all files in the test directory")
	decompressMode := flag.Bool("d", false, "Decompress mode")
	flag.Parse()

	if *inputFiles == "" {
		ColorPrint(RED, "No input files provided\n")
		flag.Usage()
		os.Exit(1)
	}

	// Read all files
	var filenameStrs []string

	// Split input files by comma and trim spaces and quotes(if any, `'` | `"`)

	if *readAllFiles {

		//filenameStrs is the folder name
		dir := inputFiles
		//if dir exists
		if _, err := os.Stat(*dir); os.IsNotExist(err) {
			ColorPrint(RED, "Directory does not exist.\n")
			os.Exit(1)
		}

		//read all filenames in the directory
		files, err := os.ReadDir(*dir)
		if err != nil {
			ColorPrint(RED, err.Error()+"\n")
			os.Exit(1)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			filenameStrs = append(filenameStrs, *dir+"/"+file.Name())
		}
	} else {
		for _, filename := range strings.Split(*inputFiles, ",") {
			filenameStrs = append(filenameStrs, strings.Trim(filename, " '\""))
		}
	}

	return filenameStrs, outputDir, password, decompressMode
}
