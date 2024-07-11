//parseCLI.go

package utils

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type MODE string

const (
	COMPRESS   MODE = "compress"
	DECOMPRESS MODE = "decompress"
)

func ParseCLI() ([]string, *string, *string, MODE) {
	// CLI arguments
	version := flag.Bool("v", false, "Print version")
	inputToCompress := flag.String("c", "", "Input files or directory to be compressed")
	outputDir := flag.String("o", "", "Output directory to compressed/decompress files (Optional)")
	password := flag.String("p", "", "Password for encryption (Optional)")
	readAllFiles := flag.Bool("a", false, "Read all files in the input directory")
	inputToDecompress := flag.String("d", "", "Input file to decompress")
	flag.Parse()
	flag.Usage = func() {
		ColorPrint(WHITE, "Usage: piarchiver [options]\n\n")
		ColorPrint(WHITE, "Options:\n")
		//print all flags in ColorPrint
		flag.VisitAll(func(f *flag.Flag) {
			ColorPrint(GREEN, fmt.Sprintf("  -%s: ", f.Name))
			ColorPrint(GREY, fmt.Sprintf("%s\n", f.Usage))
		})
	}

	if *version {
		ColorPrint(WHITE, "---------- PI ARCHIVER ----------\n")
		ColorPrint(YELLOW, "Version: v1.0.0\n")
		// dev info
		ColorPrint(WHITE, "Developed by: https://github.com/itsfuad/\n")
		ColorPrint(WHITE, "---------------------------------")
		os.Exit(0)
	}

	//mode check
	if *inputToDecompress != "" && *inputToCompress != "" {
		ColorPrint(RED, "Cannot compress and decompress at the same time\n")
		flag.Usage()
		os.Exit(1)
	}

	var filenameStrs []string
	var Mode MODE

	if *inputToCompress != "" {
		// compress mode
		Mode = COMPRESS
		// Handle reading all files in the input directory
		if *readAllFiles {
			var err error
			filenameStrs, err = GetAllFileNamesFromDir(inputToCompress)

			if err != nil {
				ColorPrint(RED, err.Error()+"\n")
				os.Exit(1)
			}
		} else {
			// Split input files by comma and trim spaces and quotes (if any)
			for _, filename := range strings.Split(*inputToCompress, ",") {
				filenameStrs = append(filenameStrs, strings.Trim(filename, " '\""))
			}
		}
	} else if *inputToDecompress != "" {
		// decompress mode
		Mode = DECOMPRESS

		//cannot contain all files lookup -a flag
		if *readAllFiles {
			ColorPrint(RED, "All files lookup not supported for decompression\n")
			flag.Usage()
			os.Exit(1)
		}

		//cannot contain comma
		if strings.Contains(*inputToDecompress, ",") {
			ColorPrint(RED, "Cannot decompress multiple files at once\n")
			flag.Usage()
			os.Exit(1)
		}

		filenameStrs = append(filenameStrs, *inputToDecompress)

	} else {
		ColorPrint(RED, "No input provided\n")
		flag.Usage()
		os.Exit(1)
	}

	return filenameStrs, outputDir, password, Mode
}

func GetAllFileNamesFromDir(dir *string) ([]string, error) {

	var filenameStrs []string

	// Check if input is a directory
	info, err := os.Stat(*dir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("input directory does not exist")
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("input is not a directory")
	}

	// Read all files in the directory
	err = filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			filenameStrs = append(filenameStrs, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return filenameStrs, nil
}
