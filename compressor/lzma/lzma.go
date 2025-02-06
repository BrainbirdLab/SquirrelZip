package lzma

import (
	"bytes"
	"encoding/binary"
	"file-compressor/constants"
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
	var inBuf bytes.Buffer
	if _, err := inBuf.ReadFrom(input); err != nil {
		return 0, err
	}
	data := inBuf.Bytes()

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
	}

	written, err := output.Write(outBuf.Bytes())
	return uint64(written), err
}

func decompressData(reader io.Reader, writer io.Writer, limiter uint64) error {
	var outBuf bytes.Buffer
	buf := make([]byte, constants.BUFFER_SIZE)
	dataRead := uint64(0)

	for {
		if limiter > 0 && dataRead >= limiter {
			break
		}

		flag, err := readFlag(reader, buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		
		dataRead++

		if err := processFlag(flag, reader, &outBuf, buf); err != nil {
			return err
		}
		dataRead += flagDataReadIncrement(flag)
	}

	_, err := writer.Write(outBuf.Bytes())
	return err
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