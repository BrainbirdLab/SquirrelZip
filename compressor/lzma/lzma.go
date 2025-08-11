package lzma

import (
	"bytes"
	"encoding/binary"
	"file-compressor/constants"
	"file-compressor/utils"
	"fmt"
	"io"
)

const (
	MaxWindowSize = 4096 // Size of sliding window
	MinMatchLen   = 3    // Minimum length for a match
)

type Match struct {
	offset int
	length int
}

func compressData(input io.Reader, output io.Writer) (uint64, error) {
	return compressDataWithProgress(input, output, "", nil)
}

// compressDataWithProgress is the enhanced version that supports progress reporting
func compressDataWithProgress(input io.Reader, output io.Writer, fileName string, progressCallback utils.ProgressCallback) (uint64, error) {
	var inBuf bytes.Buffer
	if _, err := inBuf.ReadFrom(input); err != nil {
		return 0, err
	}
	data := inBuf.Bytes()
	totalBytes := len(data)

	var outBuf bytes.Buffer
	pos := 0
	dictionary := make(map[string]int)

	for pos < len(data) {
		match := findLongestMatch(data, pos, dictionary)
		if match.length >= MinMatchLen {
			outBuf.WriteByte(1)
			binary.Write(&outBuf, binary.BigEndian, uint16(match.offset))
			outBuf.WriteByte(byte(match.length))

			for i := 0; i < match.length; i++ {
				updateDictionary(data, pos+i, dictionary)
			}
			pos += match.length
		} else {
			outBuf.WriteByte(0)
			outBuf.WriteByte(data[pos])
			updateDictionary(data, pos, dictionary)
			pos++
		}

		// Report progress every 0.01% for very smooth updates
		if progressCallback != nil && totalBytes > 0 {
			progress := float64(pos) / float64(totalBytes)
			progressCallback(progress, fmt.Sprintf("Compressing: %s (%.2f%%)", fileName, progress*100))
		}
	}

	written, err := output.Write(outBuf.Bytes())
	return uint64(written), err
}

func decompressData(reader io.Reader, writer io.Writer, limiter uint64) error {
	return decompressDataWithProgress(reader, writer, limiter, "", nil)
}

// lzmaDecompressionState holds the state during LZMA decompression
type lzmaDecompressionState struct {
	outBuf   bytes.Buffer
	dataRead uint64
}

// decompressDataWithProgress is the enhanced version that supports progress reporting
func decompressDataWithProgress(reader io.Reader, writer io.Writer, limiter uint64, fileName string, progressCallback utils.ProgressCallback) error {
	state := &lzmaDecompressionState{}
	buf := make([]byte, constants.BUFFER_SIZE)

	for {
		if shouldStopDecompression(state.dataRead, limiter) {
			break
		}

		if err := processDecompressionStep(reader, &state.outBuf, buf, state, limiter, fileName, progressCallback); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	_, err := writer.Write(state.outBuf.Bytes())
	return err
}

// shouldStopDecompression checks if decompression should stop
func shouldStopDecompression(dataRead, limiter uint64) bool {
	return limiter > 0 && dataRead >= limiter
}

// processDecompressionStep processes a single step in LZMA decompression
func processDecompressionStep(reader io.Reader, outBuf *bytes.Buffer, buf []byte, state *lzmaDecompressionState, limiter uint64, fileName string, progressCallback utils.ProgressCallback) error {
	flag, err := readFlag(reader, buf)
	if err != nil {
		return err
	}

	state.dataRead++

	if err := processFlag(flag, reader, outBuf, buf); err != nil {
		return err
	}
	state.dataRead += flagDataReadIncrement(flag)

	// Report progress during decompression
	reportLZMADecompressionProgress(progressCallback, state.dataRead, limiter, fileName)

	return nil
}

// reportLZMADecompressionProgress reports the current LZMA decompression progress
func reportLZMADecompressionProgress(progressCallback utils.ProgressCallback, dataRead, limiter uint64, fileName string) {
	if progressCallback != nil && limiter > 0 {
		progress := float64(dataRead) / float64(limiter)
		if progress > 1.0 {
			progress = 1.0
		}
		progressCallback(progress, fmt.Sprintf("Decompressing: %s (%.2f%%)", fileName, progress*100))
	}
}

func processFlag(flag byte, reader io.Reader, outBuf *bytes.Buffer, buf []byte) error {
	if flag == 0 {
		return handleLiteral(reader, outBuf, buf)
	}
	return handleMatch(reader, outBuf, buf)
}

func flagDataReadIncrement(flag byte) uint64 {
	if flag == 0 {
		return 1
	}
	return 3
}

func readFlag(reader io.Reader, buf []byte) (byte, error) {
	_, err := io.ReadFull(reader, buf[:1])
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func handleLiteral(reader io.Reader, outBuf *bytes.Buffer, buf []byte) error {
	_, err := io.ReadFull(reader, buf[:1])
	if err != nil {
		return err
	}
	outBuf.WriteByte(buf[0])
	return nil
}

func handleMatch(reader io.Reader, outBuf *bytes.Buffer, buf []byte) error {
	_, err := io.ReadFull(reader, buf[:3])
	if err != nil {
		return err
	}

	offset := int(binary.BigEndian.Uint16(buf[:2]))
	length := int(buf[2])

	if offset > outBuf.Len() {
		return fmt.Errorf("invalid offset")
	}
	startPos := outBuf.Len() - offset
	for i := 0; i < length; i++ {
		outBuf.WriteByte(outBuf.Bytes()[startPos+i])
	}
	return nil
}

func findLongestMatch(data []byte, pos int, dict map[string]int) Match {
	if pos >= len(data) {
		return Match{0, 0}
	}

	maxLen := MinMatchLen - 1
	bestMatch := Match{0, 0}

	for length := MinMatchLen; length <= MaxWindowSize && pos+length <= len(data); length++ {
		searchBytes := data[pos : pos+length]
		if prevPos, exists := dict[string(searchBytes)]; exists {
			offset := pos - prevPos
			if offset <= MaxWindowSize && length > maxLen {
				maxLen = length
				bestMatch = Match{offset, length}
			}
		}
	}

	return bestMatch
}

func updateDictionary(data []byte, pos int, dict map[string]int) {
	maxSubstringLen := MinMatchLen * 2
	for length := MinMatchLen; length <= maxSubstringLen && pos+length <= len(data); length++ {
		substr := string(data[pos : pos+length])
		dict[substr] = pos
	}
}
