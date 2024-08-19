package main

import (
	"time"
	
	"file-compressor/compressor"
	"file-compressor/utils"
)



func main() {
	//test files path '/test'

	startTime :=  time.Now()

	//cli arguments
	filenameStrs, outputDir, password, mode, algorithm := utils.ParseCLI()

	var err error

	if mode == utils.DECOMPRESS {
		err = compressor.Decompress(filenameStrs[0], *outputDir, *password)
	} else {
		err = compressor.Compress(filenameStrs, *outputDir, *password, algorithm)
	}

	if err != nil {
		utils.ColorPrint(utils.RED, err.Error() + "\n")
	} else {
		endTime := time.Now()
		utils.ColorPrint(utils.GREEN, "Time taken: " + utils.TimeTrack(startTime, endTime) + "\n")
	}
}
