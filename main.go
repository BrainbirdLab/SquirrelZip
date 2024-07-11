// go: generate goversioninfo -icon = icon_piarchiver.ico

package main

import (
	"file-compressor/compressor"
	"file-compressor/utils"
	"time"
	"fmt"
)


func main() {
	//test files path '/test'

	startTime :=  time.Now()

	//cli arguments
	filenameStrs, outputDir, password, mode := utils.ParseCLI()

	var err error

	if mode == utils.DECOMPRESS {
		err = compressor.Decompress(filenameStrs[0], *outputDir, *password)
	} else {
		err = compressor.Compress(filenameStrs, *outputDir, *password)
	}

	if err != nil {
		utils.ColorPrint(utils.RED, err.Error() + "\n")
	} else {
		endTime := time.Now()
		tookTime := endTime.Sub(startTime).Microseconds()
		utils.ColorPrint(utils.GREEN, fmt.Sprintf("Took %d ms\n", tookTime))
	}
}
