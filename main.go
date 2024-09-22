package main

import (
	"file-compressor/compressor"
	"file-compressor/utils"
	"os"
	"time"
)


func main() {

	startTime :=  time.Now()

	//cli arguments
	filenameStrs, outputDir, password, mode, algorithm := utils.ParseCLI()

	if mode == utils.DECOMPRESS {
		paths, err := compressor.Decompress(filenameStrs[0], outputDir, password)
		if err != nil {
			utils.ColorPrint(utils.RED, err.Error() + "\n")
			os.Exit(-1)
		}

		for _, path := range paths {
			utils.ColorPrint(utils.GREEN, "Decompressed file: " + path + "\n")
		}

	} else {
		outputPath, fileMeta, err := compressor.Compress(filenameStrs, outputDir, password, algorithm)
		if err != nil {
			utils.ColorPrint(utils.RED, err.Error() + "\n")
			os.Exit(-1)
		}

		fileMeta.PrintFileInfo()
		fileMeta.PrintCompressionRatio()

		utils.ColorPrint(utils.GREEN, "Compressed file: " + outputPath + "\n")
	}

	endTime := time.Now()
	utils.ColorPrint(utils.GREEN, "Time taken: " + utils.TimeTrack(startTime, endTime) + "\n")
}
