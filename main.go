// go: generate goversioninfo -icon = icon_piarchiver.ico

package main

import (
	"file-compressor/compressor"
	"file-compressor/utils"
	"os"
)


func main() {
	//test files path '/test'

	//cli arguments
	filenameStrs, outputDir, password, decompressMode := utils.ParseCLI()

	var err error

	if *decompressMode {
		err = compressor.Decompress(filenameStrs, outputDir, password)
	} else {
		err = compressor.Compress(filenameStrs, outputDir, password)
	}

	if err != nil {
		utils.ColorPrint(utils.RED, err.Error() + "\n")
		os.Exit(1)
	}
}
