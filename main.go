package main

import (
	"time"
	
	"file-compressor/utils"
)



func main() {
	//test files path '/test'

	startTime :=  time.Now()

	var err error


	if err != nil {
		utils.ColorPrint(utils.RED, err.Error() + "\n")
	} else {
		endTime := time.Now()
		utils.ColorPrint(utils.GREEN, "Time taken: " + utils.TimeTrack(startTime, endTime) + "\n")
	}
}
