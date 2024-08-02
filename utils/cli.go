//parseCLI.go

package utils

import (
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

type FlagSet struct {
	flags       map[string]*Flag
	parsedFlags map[string]interface{}
}

type Flag struct {
	Name    string
	Usage   string
	IsBool  bool
	IsArray bool
}

func NewFlagSet() *FlagSet {
	return &FlagSet{
		flags:       make(map[string]*Flag),
		parsedFlags: make(map[string]interface{}),
	}
}

func (fs *FlagSet) Bool(name, usage string) {
	fs.flags[name] = &Flag{Name: name, Usage: usage, IsBool: true}
	fs.parsedFlags[name] = false
}

func (fs *FlagSet) String(name, usage string) {
	fs.flags[name] = &Flag{Name: name, Usage: usage, IsBool: false}
}

func (fs *FlagSet) ArrayStr(name, usage string) {
	fs.flags[name] = &Flag{Name: name, Usage: usage, IsBool: false, IsArray: true}
	fs.parsedFlags[name] = []string{}
}

func (fs *FlagSet) Parse(args []string) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if err := fs.processArg(arg, &i, args); err != nil {
			return err
		}
	}
	return nil
}

func (fs *FlagSet) processArg(arg string, i *int, args []string) error {
	if !strings.HasPrefix(arg, "-") {
		return fmt.Errorf("invalid argument: %s", arg)
	}
	flagName := arg[1:]
	flag, exists := fs.flags[flagName]
	if !exists {
		return fmt.Errorf("unknown flag: %s", flagName)
	}
	if flag.IsBool {
		fs.parsedFlags[flag.Name] = true
	} else if flag.IsArray {
		return fs.collectArrayValues(flagName, i, args)
	} else {
		return fs.collectValues(flagName, i, args)
	}
	return nil
}

func (fs *FlagSet) collectArrayValues(flagName string, i *int, args []string) error {
	values := []string{}
	for j := *i + 1; j < len(args); j++ {
		if strings.HasPrefix(args[j], "-") {
			break
		}
		values = append(values, args[j])
		*i = j
	}
	if len(values) == 0 {
		return fmt.Errorf("flag -%s requires a value", flagName)
	}
	fs.parsedFlags[flagName] = values
	return nil
}

func (fs *FlagSet) collectValues(flagName string, i *int, args []string) error {
	if *i+1 >= len(args) || strings.HasPrefix(args[*i+1], "-") {
		return fmt.Errorf("flag -%s requires a value", flagName)
	}
	fs.parsedFlags[flagName] = args[*i+1]
	*i++
	return nil
}

func (fs *FlagSet) Get(flagName string) (interface{}, bool) {
	value, exists := fs.parsedFlags[flagName]
	return value, exists
}

func (fs *FlagSet) Usage() {
	fmt.Println("Usage: Chipmunk file archiver [options]")
	fmt.Println("Options:")
	for _, flag := range fs.flags {
		fmt.Printf("  -%s: %s\n", flag.Name, flag.Usage)
	}
}

var flagSet = NewFlagSet()

func initFlags() (map[string]interface{}, error) {
	flagSet.Bool("v", "Print version")
	flagSet.ArrayStr("c", "Input files or directory to be compressed")
	flagSet.String("o", "Output directory to compressed/decompress files (Optional)")
	flagSet.String("p", "Password for encryption (Optional)")
	flagSet.Bool("a", "Read all files in the input directory")
	flagSet.ArrayStr("d", "Input file to decompress")
	flagSet.Bool("h", "Print help")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		flagSet.Usage()
		return nil, err
	}

	return flagSet.parsedFlags, nil
}

func setupCompressMode(Mode *MODE, readAllFiles *bool, inputToCompress []string, filenameStrs *[]string) {
	// compress mode
	*Mode = COMPRESS
	// Handle reading all files in the input directory
	if *readAllFiles {
		var err error

		*filenameStrs, err = GetAllFileNamesFromDir(&inputToCompress[0])

		if err != nil {
			ColorPrint(RED, err.Error()+"\n")
			os.Exit(1)
		}
	} else {
		*filenameStrs = inputToCompress
	}
}

func setupDecompressMode(Mode *MODE, inputToDecompress []string, filenameStrs *[]string, readAllFiles *bool) {
	// decompress mode
	*Mode = DECOMPRESS

	//cannot contain all files lookup -a flag
	if *readAllFiles {
		ColorPrint(RED, "All files lookup not supported for decompression\n")
		flagSet.Usage()
		os.Exit(1)
	}

	//cannot contain comma
	if  len(inputToDecompress) > 1 {
		ColorPrint(RED, "Cannot decompress multiple files at once\n")
		flagSet.Usage()
		os.Exit(1)
	}

	*filenameStrs = append(*filenameStrs, inputToDecompress[0])
}

func ParseCLI() ([]string, *string, *string, MODE) {
	// CLI arguments

	values, err := initFlags()

	if err != nil {
		ColorPrint(RED, err.Error()+"\n")
		os.Exit(1)
	}

	// flags
	help, _ := values["h"].(bool)
	version, _ := values["v"].(bool)
	inputToCompress, _ := values["c"].([]string)
	outputDir, _ := values["o"].(string)
	password, _ := values["p"].(string)
	readAllFiles, _ := values["a"].(bool)
	inputToDecompress, _ := values["d"].([]string)

	if version {
		ColorPrint(WHITE, "---------- SquirrelZip ----------\n")
		ColorPrint(YELLOW, "Version: v1.0.8\n")
		// dev info
		ColorPrint(WHITE, "Developed by: https://github.com/itsfuad/\n")
		ColorPrint(WHITE, "---------------------------------")
		os.Exit(0)
	}

	if help {
		flagSet.Usage()
		os.Exit(0)
	}

	//mode check
	if len(inputToDecompress) > 0 && len(inputToCompress) > 0 {
		ColorPrint(RED, "Cannot compress and decompress at the same time\n")
		flagSet.Usage()
		os.Exit(1)
	}

	var filenameStrs []string
	var Mode MODE

	if len(inputToCompress) > 0 {
		setupCompressMode(&Mode, &readAllFiles, inputToCompress, &filenameStrs)
	} else if len(inputToCompress) == 0 && len(inputToDecompress) == 0 {
		ColorPrint(RED, "No input files provided\n")
		flagSet.Usage()
		os.Exit(1)
	} else if len(inputToDecompress) > 0 {
		setupDecompressMode(&Mode, inputToDecompress, &filenameStrs, &readAllFiles)
	} else {
		ColorPrint(RED, "No flags provided\n")
		flagSet.Usage()
		os.Exit(1)
	}

	return filenameStrs, &outputDir, &password, Mode
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
