package lz77

import (
	"file-compressor/utils"
	"fmt"
	"os"
	"testing"
)

func TestLZ(t *testing.T) {

	compressor := NewLZ77Compressor()

	targetData := "Hello world Hello world"
	compressedData, err := compressor.CompressData([]byte(targetData))
	if err != nil {
		t.Fatalf(utils.FailedToCompress, err)
	}
	
	fileRatio := utils.NewFilesRatio(len(targetData), len(compressedData))

	fileRatio.PrintFileInfo()

	decompressed, err := compressor.DecompressData(compressedData)
	if err != nil {
		t.Fatalf(utils.FailedToCompress, err)
	}
	if string(decompressed) != targetData {
		t.Fatalf(utils.MissMatch, targetData, decompressed)
	}

	fileRatio.PrintCompressionRatio()
}

func TestSmallData(t *testing.T) {
	compressor := NewLZ77Compressor()
	smallData := "ababcabcabcd"
	compressed, err := compressor.CompressData([]byte(smallData))
	if err != nil {
		t.Fatalf(utils.FailedToCompress, err)
	}

	decompressed, err := compressor.DecompressData(compressed)
	if err != nil {
		t.Fatalf(utils.FailedToDecompress, err)
	}
	if string(decompressed) != smallData {
		t.Fatalf(utils.MissMatch, smallData, decompressed)
	}
}


func TestLZWithFileData(t *testing.T) {
	RunFile("example.txt", t)
	RunFile("image.JPG", t)
}

func RunFile(targetPath string, t *testing.T) {
	_, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("failed to get target file info: %v", err)
	}

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed to read target file: %v", err)
	}

	compressor := NewLZ77Compressor()

	fmt.Printf("Compressing %s\n", targetPath)

	compressedData, err := compressor.CompressData(targetData)
	if err != nil {
		t.Fatalf(utils.FailedToCompress, err)
	}

	fileRatio := utils.NewFilesRatio(len(targetData), len(compressedData))

	fileRatio.PrintFileInfo()

	decompressedData, err := compressor.DecompressData(compressedData)
	if err != nil {
		t.Fatalf(utils.FailedToDecompress, err)
	}
	if string(decompressedData) != string(targetData) {
		t.Fatalf(utils.MissMatch, targetData, decompressedData)
	}

	fileRatio.PrintCompressionRatio()
}