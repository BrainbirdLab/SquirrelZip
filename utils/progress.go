package utils

import (
	"fmt"
	"path/filepath"
)

// ProgressCallback is a function type that will be called to report progress
type ProgressCallback func(progress float64, message string)

// ProgressInfo contains information about the current operation for progress reporting
type ProgressInfo struct {
	CurrentFileName     string // Name of the file currently being processed
	SizeOfCurrentFile   int64  // Total size of the current file
	BytesProcessedForCurrentFile int64 // Bytes processed for the current file in the current operation

	TotalFilesOverall     int   // Total number of files to process in the batch
	FilesCompletedOverall int   // Number of files fully processed so far in the batch

	OverallTotalBytes    int64 // Total size of all files in the batch
	AccumulatedBytesProcessedFromPreviousFiles int64 // Total bytes from previously completed files in this batch

	OriginalMessage string // Optional: original message from the underlying process
}

// UpdateProgress calculates and calls the progress callback with the current progress
func UpdateProgress(info ProgressInfo, callback ProgressCallback) {
	if callback == nil {
		return
	}

	if info.OverallTotalBytes == 0 { // Avoid division by zero; indicates no data or start
		// Try to provide a meaningful start message even if CurrentFileName is not yet set
		startMsg := "Starting operation..."
		if info.CurrentFileName != "" {
			startMsg = fmt.Sprintf("Starting operation on %s...", filepath.Base(info.CurrentFileName))
		} else if info.OriginalMessage != "" {
            startMsg = info.OriginalMessage
        }
		callback(0, startMsg)
		return
	}

	// Calculate overall progress based on bytes
	totalBytesEffectivelyProcessed := info.AccumulatedBytesProcessedFromPreviousFiles + info.BytesProcessedForCurrentFile
	overallProgress := float64(totalBytesEffectivelyProcessed) / float64(info.OverallTotalBytes)

	// Ensure progress is between 0 and 1
	if overallProgress < 0 {
		overallProgress = 0
	}
	if overallProgress > 1 {
		overallProgress = 1
	}

	// Construct a detailed message
	// File X/Y: filename (current_file_bytes/current_file_total_bytes) original_message
	var progressMessage string
	if info.CurrentFileName != "" { // Only show file-specific info if CurrentFileName is available
		if info.TotalFilesOverall > 0 {
			progressMessage = fmt.Sprintf("File %d/%d: %s",
										  info.FilesCompletedOverall + 1, // Current file is "+1" to completed
										  info.TotalFilesOverall,
										  filepath.Base(info.CurrentFileName))
		} else {
			progressMessage = fmt.Sprintf("Processing: %s", filepath.Base(info.CurrentFileName))
		}

		if info.SizeOfCurrentFile > 0 {
			progressMessage += fmt.Sprintf(" (%s / %s)",
										   FileSize(uint64(info.BytesProcessedForCurrentFile)),
										   FileSize(uint64(info.SizeOfCurrentFile)))
		}
	}

	if info.OriginalMessage != "" {
		if progressMessage == "" { // If no file-specific message part, just use original
			progressMessage = info.OriginalMessage
		} else {
			progressMessage += " - " + info.OriginalMessage
		}
	} else if progressMessage == "" { // Fallback if no info at all
        progressMessage = fmt.Sprintf("Progress: %.0f%%", overallProgress*100)
    }


	callback(overallProgress, progressMessage)
}
