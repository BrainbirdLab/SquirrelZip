package lampelziv

import (
	"file-compressor/utils"
	"os"
	"testing"
)

func TestLZ(t *testing.T) {
	targetData := "Hello world Hello world"
	compressedData, err := CompressData([]byte(targetData))
	if err != nil {
		t.Fatalf(utils.FailedToCompress, err)
	}
	
	fileRatio := utils.NewFilesRatio(len(targetData), len(compressedData))

	fileRatio.PrintFileInfo()

	decompressed, err := DecompressData(compressedData)
	if err != nil {
		t.Fatalf(utils.FailedToCompress, err)
	}

	if string(decompressed) != targetData {
		t.Fatalf(utils.MissMatch, targetData, decompressed)
	}

	fileRatio.PrintCompressionRatio()
}

func TestSmallData(t *testing.T) {
	targetData := "ababcabcabcd"
	compressedData, err := CompressData([]byte(targetData))
	if err != nil {
		t.Fatalf("failed to compress data: %v", err)
	}

	fileRatio := utils.NewFilesRatio(len(targetData), len(compressedData))

	fileRatio.PrintFileInfo()

	decompressed, err := DecompressData(compressedData)
	if err != nil {
		t.Fatalf("failed to decompress data: %v", err)
	}
	if string(decompressed) != targetData {
		t.Fatalf(utils.MissMatch, targetData, decompressed)
	}

	fileRatio.PrintCompressionRatio()
}


func TestLZWithFileData(t *testing.T) {
	
	targetPath := "example.txt"

	_, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("failed to get target file info: %v", err)
	}

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed to read target file: %v", err)
	}

	compressedData, err := CompressData(targetData)
	if err != nil {
		t.Fatalf(utils.FailedToCompress, err)
	}

	fileRatio := utils.NewFilesRatio(len(targetData), len(compressedData))

	fileRatio.PrintFileInfo()

	decompressedData, err := DecompressData(compressedData)
	if err != nil {
		t.Fatalf("failed to decompress data: %v", err)
	}
	if string(decompressedData) != string(targetData) {
		t.Fatalf("decompressed data does not match original data: %v != %v", targetData, decompressedData)
	}

	fileRatio.PrintCompressionRatio()
}