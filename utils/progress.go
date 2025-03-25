package utils

// ProgressCallback is a function type that will be called to report progress
type ProgressCallback func(progress float64, message string)

// ProgressInfo contains information about the current operation
type ProgressInfo struct {
	TotalFiles      int
	CurrentFile     int
	CurrentFileSize int64
	ProcessedBytes  int64
	Message         string
}

// UpdateProgress calculates and calls the progress callback with the current progress
func UpdateProgress(info ProgressInfo, callback ProgressCallback) {
	if callback == nil {
		return
	}

	// Handle initial state where values might be 0 or invalid
	if info.TotalFiles == 0 || info.CurrentFileSize == 0 {
		callback(0, info.Message)
		return
	}

	// Calculate overall progress
	fileProgress := float64(info.CurrentFile) / float64(info.TotalFiles)
	byteProgress := float64(info.ProcessedBytes) / float64(info.CurrentFileSize)

	// Combine both progress indicators
	progress := (fileProgress + byteProgress) / 2.0

	// Ensure progress is between 0 and 1
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	callback(progress, info.Message)
}
