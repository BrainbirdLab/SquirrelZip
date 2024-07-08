package main

import (
	"file-compressor/compressor"
	"file-compressor/utils"
)


func main() {
	//test files path '/test'

	//cli arguments
	filenameStrs, outputDir, decompressMode := utils.ParseCLI()

	if *decompressMode {
		compressor.Decompress(filenameStrs, outputDir)
	} else {
		compressor.Compress(filenameStrs, outputDir)
	}

	utils.ColorPrint(utils.GREEN, "Done\n")
}
