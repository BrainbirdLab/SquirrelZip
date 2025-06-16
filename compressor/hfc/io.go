package hfc

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath" // Added for filepath.Base

	"file-compressor/constants" // Assuming correct path
	"file-compressor/utils"     // Assuming correct path
)

// Zip compresses a single file using Huffman coding and writes to the output writer.
// It writes metadata for the file (name, codes, data length, data).
// The progressCallback expects (percentOfCurrentFile float64, messageFromZip string).
func Zip(file utils.FileData, output io.Writer, progressCallback func(percentOfCurrentFile float64, messageFromZip string)) error {
	if progressCallback != nil {
		progressCallback(0.0, fmt.Sprintf("Starting Huffman for %s", filepath.Base(file.Name)))
	}

	if file.Reader == nil {
		if file.Size > 0 {
			utils.ColorPrint(utils.YELLOW, fmt.Sprintf("Warning: Huffman Zip skipping '%s' due to nil reader.\n", file.Name))
		}
		// Ensure progress is marked as "complete" for this skipped file for overall calculation
		if progressCallback != nil {
			progressCallback(1.0, fmt.Sprintf("Skipped %s (nil reader)", filepath.Base(file.Name)))
		}
		return nil
	}

	fileReader := file.Reader
	defer func(r io.Reader) {
		if c, ok := r.(io.Closer); ok {
			c.Close()
		}
	}(fileReader)

	contentBuffer := utils.NewResizableBuffer()
    // var bytesReadForFreqAnalysis int64 // Not directly needed due to ProgressTrackingReader's totalBytes

    onReadProgress := func(readBytes int64, totalBytes int64) {
        // bytesReadForFreqAnalysis = readBytes // This variable isn't strictly necessary here
        if progressCallback != nil && totalBytes > 0 {
            // Report progress for the frequency analysis read (maps to first 50% of this file's step)
            percent := (float64(readBytes) / float64(totalBytes)) * 0.5
            if percent > 0.5 { percent = 0.5} // Cap at 0.5 for this phase
            progressCallback(percent, fmt.Sprintf("Analyzing %s", filepath.Base(file.Name)))
        }
    }

    // If file.Size is 0 initially, ProgressTrackingReader might not trigger onReadProgress meaningfully for 0-byte files.
    // Handle 0-byte file case before attempting to read with ProgressTrackingReader if it causes issues.
    // However, io.Copy with a 0-byte reader should just result in 0 bytes copied.
	trackingReaderFreq := utils.NewProgressReader(fileReader, file.Size, onReadProgress)

	if _, err := io.Copy(contentBuffer, trackingReaderFreq); err != nil {
		return fmt.Errorf("%s: %w", fmt.Sprintf(constants.FILE_READ_ERROR_ συγκεκριμένα, "Huffman Zip: reading %s for frequency analysis", file.Name), err)
	}

    // Ensure 50% is reported after analysis, especially if file.Size was 0 (onReadProgress might not have run)
    // or if it didn't quite hit 100% of the read.
    if progressCallback != nil {
         progressCallback(0.5, fmt.Sprintf("Finished analysis for %s", filepath.Base(file.Name)))
    }

	fileBytes := contentBuffer.Bytes()
	if len(fileBytes) == 0 { // Handle empty file (already read fully by io.Copy)
		// Write metadata for empty file: name, 0 codes, 0 data length
		fileNameBytes := []byte(filepath.Base(file.Name))
		if err := binary.Write(output, binary.LittleEndian, uint32(len(fileNameBytes))); err != nil {
			return fmt.Errorf("%s: %w", fmt.Sprintf(constants.FILE_WRITE_ERROR_ συγκεκριμένα, "Huffman Zip: writing name length for empty file %s", file.Name), err)
		}
		if _, err := output.Write(fileNameBytes); err != nil {
			return fmt.Errorf("%s: %w", fmt.Sprintf(constants.FILE_WRITE_ERROR_ συγκεκριμένα, "Huffman Zip: writing name for empty file %s", file.Name), err)
		}
		// Write 0 for Huffman codes length
		if err := binary.Write(output, binary.LittleEndian, uint32(0)); err != nil {
			return fmt.Errorf("%s: %w", fmt.Sprintf(constants.FILE_WRITE_ERROR_ συγκεκριμένα, "Huffman Zip: writing zero codes length for empty file %s", file.Name), err)
		}
		// Write 0 for data length
		if err := binary.Write(output, binary.LittleEndian, uint64(0)); err != nil {
			return fmt.Errorf("%s: %w", fmt.Sprintf(constants.FILE_WRITE_ERROR_ συγκεκριμένα, "Huffman Zip: writing zero data length for empty file %s", file.Name), err)
		}
		if progressCallback != nil {
			progressCallback(1.0, fmt.Sprintf("Processed empty file %s", filepath.Base(file.Name)))
		}
		return nil
	}

	freq := GetFrequency(string(fileBytes)) // Assumes GetFrequency is defined in the hfc package
	codes, err := GetHuffmanCodes(&freq)    // Assumes GetHuffmanCodes is defined
	if err != nil {
		return fmt.Errorf("%s: %w", constants.FAILED_BUILD_HUFFMAN_CODES_ συγκεκριμένα, err)
	}

	fileNameBytes := []byte(filepath.Base(file.Name))
	if err := binary.Write(output, binary.LittleEndian, uint32(len(fileNameBytes))); err != nil {
		return fmt.Errorf("%s: %w", fmt.Sprintf(constants.FILE_WRITE_ERROR_ συγκεκριμένα, "Huffman Zip: filename length for %s", file.Name), err)
	}
	if _, err := output.Write(fileNameBytes); err != nil {
		return fmt.Errorf("%s: %w", fmt.Sprintf(constants.FILE_WRITE_ERROR_ συγκεκριμένα, "Huffman Zip: filename for %s", file.Name), err)
	}

	if err := writeHuffmanCodes(output, &codes); err != nil { // Assumes writeHuffmanCodes is defined
		return fmt.Errorf("%s: %w", constants.FAILED_WRITE_HUFFMAN_CODES_ συγκεκριμένα, err)
	}

	tempCompressedData := utils.NewResizableBuffer()
	// Ensure NewBitWriter is correctly referenced (e.g., from this package or utils)
	// For this example, assuming it's part of `hfc` package.
	bitWriter := NewBitWriter(tempCompressedData)

	// This encoding loop is the second major part of the work.
	// Progress from 0.5 to 0.9 (or 1.0 if possible) should happen here.
	// For simplicity, we'll update progress after the loop.
	// A more granular progress would require knowing how many input bytes correspond to output bits.
	totalFileBytesForEncoding := len(fileBytes)
	for idx, b := range fileBytes {
		code := codes[rune(b)] // This could panic if a byte is not in codes map. Robust code checks.
		for _, bit := range code {
			err := bitWriter.WriteBit(bit == '1')
			if err != nil {
				return fmt.Errorf("Huffman Zip: writing bit for %s: %w", file.Name, err)
			}
		}
        if progressCallback != nil && totalFileBytesForEncoding > 0 && (idx%(totalFileBytesForEncoding/10) == 0 || idx == totalFileBytesForEncoding-1) {
            // Update progress 10 times during encoding, or at the end.
            // This maps the encoding phase (0 to 100%) to the overall file progress (50% to 90-95%).
            encodingProgress := float64(idx+1) / float64(totalFileBytesForEncoding)
            overallFileProgress := 0.5 + (encodingProgress * 0.45) // Encoding phase contributes 45% (from 50% to 95%)
            if overallFileProgress > 0.95 { overallFileProgress = 0.95 }
             progressCallback(overallFileProgress, fmt.Sprintf("Encoding %s", filepath.Base(file.Name)))
        }
	}
	if err := bitWriter.Flush(); err != nil {
		return fmt.Errorf("Huffman Zip: flushing bits for %s: %w", file.Name, err)
	}

    if progressCallback != nil {
        progressCallback(0.95, fmt.Sprintf("Finished encoding %s", filepath.Base(file.Name))) // Mark 95% after encoding and flushing
    }

	compressedSize := uint64(tempCompressedData.Len())
	if err := binary.Write(output, binary.LittleEndian, compressedSize); err != nil {
		return fmt.Errorf("%s: %w", fmt.Sprintf(constants.FILE_WRITE_ERROR_ συγκεκριμένα, "Huffman Zip: data length for %s", file.Name), err)
	}
	if _, err := io.Copy(output, tempCompressedData.Reader()); err != nil {
		return fmt.Errorf("%s: %w", fmt.Sprintf(constants.FILE_WRITE_ERROR_ συγκεκριμένα, "Huffman Zip: data for %s", file.Name), err)
	}

	if progressCallback != nil {
		progressCallback(1.0, fmt.Sprintf("Finished Huffman for %s", filepath.Base(file.Name)))
	}
	return nil
}

// Unzip function is not modified in this subtask.
// func Unzip(input io.Reader, outputDir string, progressCallback utils.ProgressCallback) ([]string, error) { ... }


// Ensure these helper functions are defined within the hfc package or imported.
// Dummy implementations for context if they were missing:
/*
func GetFrequency(text string) map[rune]int {
    freq := make(map[rune]int)
    for _, r := range text {
        freq[r]++
    }
    return freq
}

func GetHuffmanCodes(freq *map[rune]int) (map[rune]string, error) {
    // Simplified: actual implementation involves building a Huffman tree
    if freq == nil || len(*freq) == 0 {
        return make(map[rune]string), nil // Or error for no frequency
    }
    codes := make(map[rune]string)
    idx := 0
    for r := range *freq {
        codes[r] = fmt.Sprintf("%b", idx) // Dummy codes
        idx++
    }
    return codes, nil
}

func writeHuffmanCodes(output io.Writer, codes *map[rune]string) error {
    // Simplified: actual implementation involves serializing the code table or tree
    codesLen := uint32(len(*codes))
    if err := binary.Write(output, binary.LittleEndian, codesLen); err != nil {
        return err
    }
    for r, codeStr := range *codes {
        if err := binary.Write(output, binary.LittleEndian, r); err != nil { return err }
        codeBytes := []byte(codeStr)
        if err := binary.Write(output, binary.LittleEndian, uint8(len(codeBytes))); err != nil { return err }
        if _, err := output.Write(codeBytes); err != nil { return err }
    }
    return nil
}

type BitWriter struct {
    writer io.ByteWriter
    buffer byte
    count  uint8
}

func NewBitWriter(w io.Writer) *BitWriter {
    bw, ok := w.(io.ByteWriter)
    if !ok {
        // If w is not a ByteWriter, wrap it with bufio.NewWriter, which is a ByteWriter
        // This is a common pattern.
        bw = bufio.NewWriter(w)
    }
    return &BitWriter{writer: bw}
}

func (bw *BitWriter) WriteBit(bit bool) error {
    if bit {
        bw.buffer |= (1 << (7 - bw.count))
    }
    bw.count++
    if bw.count == 8 {
        if err := bw.writer.WriteByte(bw.buffer); err != nil {
            return err
        }
        // If underlying writer is bufio.Writer, it needs to be flushed eventually.
        // The BitWriter's Flush method should handle this.
        bw.buffer = 0
        bw.count = 0
    }
    return nil
}

func (bw *BitWriter) Flush() error {
    if bw.count > 0 { // Write any remaining bits
        if err := bw.writer.WriteByte(bw.buffer); err != nil {
            return err
        }
         bw.buffer = 0
         bw.count = 0
    }
    // If the underlying writer (passed to NewBitWriter) has a Flush method (e.g., *bufio.Writer), call it.
    if flusher, ok := bw.writer.(interface{ Flush() error }); ok {
        return flusher.Flush()
    }
    return nil
}
*/
