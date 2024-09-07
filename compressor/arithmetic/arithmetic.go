package arithmetic

import (
	"bytes"
	"encoding/binary"
	"file-compressor/utils"
	"file-compressor/encoder"
	"math"
)

type ProbabilityRange struct {
	low  float64
	high float64
}

func calculateFrequencies(data []byte) map[byte]int {
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}
	return freq
}

func calculateProbabilityRanges(freq map[byte]int) map[byte]ProbabilityRange {
	total := 0
	for _, f := range freq {
		total += f
	}

	ranges := make(map[byte]ProbabilityRange)
	low := 0.0

	for char, f := range freq {
		high := low + float64(f)/float64(total)
		ranges[char] = ProbabilityRange{low, high}
		low = high
	}
	return ranges
}

func compressData(data []byte, ranges map[byte]ProbabilityRange) (float64, int) {
	low, high := 0.0, 1.0
	for _, b := range data {
		r := ranges[b]
		rangeWidth := high - low
		high = low + rangeWidth*r.high
		low = low + rangeWidth*r.low
	}

	// To encode the range, choose a number in the final interval
	code := (low + high) / 2
	return code, len(data)
}

func encodeBinary(code float64, numBits int) []byte {
	var buf bytes.Buffer
	for i := 0; i < numBits; i++ {
		code *= 2
		if code >= 1.0 {
			buf.WriteByte(1)
			code -= 1.0
		} else {
			buf.WriteByte(0)
		}
	}
	return buf.Bytes()
}

func Zip(files []utils.FileData) (utils.FileData, error) {
	var buf bytes.Buffer

	err := binary.Write(&buf, binary.BigEndian, uint32(len(files)))
	if err != nil {
		return utils.FileData{}, err
	}

	var rawContent bytes.Buffer
	err = encoder.CreateContentBuffer(files, &rawContent)
	if err != nil {
		return utils.FileData{}, err
	}

	freq := calculateFrequencies(rawContent.Bytes())
	ranges := calculateProbabilityRanges(freq)
	code, numBits := compressData(rawContent.Bytes(), ranges)

	encodedData := encodeBinary(code, numBits)
	err = binary.Write(&buf, binary.BigEndian, uint32(len(encodedData)))
	if err != nil {
		return utils.FileData{}, err
	}
	_, err = buf.Write(encodedData)
	if err != nil {
		return utils.FileData{}, err
	}

	err = writeProbabilityRangesToBuffer(ranges, &buf)
	if err != nil {
		return utils.FileData{}, err
	}

	return utils.FileData{
		Name:    "compressed.sq",
	}, nil
}

func Unzip(file utils.FileData) ([]utils.FileData, error) {
	var files []utils.FileData
	buf := bytes.NewBuffer([]byte(""))

	var numFiles uint32
	err := binary.Read(buf, binary.BigEndian, &numFiles)
	if err != nil {
		return nil, err
	}

	var encodedDataLength uint32
	err = binary.Read(buf, binary.BigEndian, &encodedDataLength)
	if err != nil {
		return nil, err
	}

	encodedData := make([]byte, encodedDataLength)
	_, err = buf.Read(encodedData)
	if err != nil {
		return nil, err
	}

	ranges := make(map[byte]ProbabilityRange)
	err = readProbabilityRanges(&ranges, buf)
	if err != nil {
		return nil, err
	}

	// The last argument is the length of the original data
	decodedData := decodeArithmetic(encodedData, ranges, int(encodedDataLength))

	decompressedContentBuf := bytes.NewBuffer(decodedData)
	err = encoder.ParseDecompressedContent(&files, &numFiles, decompressedContentBuf)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func writeProbabilityRangesToBuffer(ranges map[byte]ProbabilityRange, buf *bytes.Buffer) error {
	for char, r := range ranges {
		err := binary.Write(buf, binary.BigEndian, char)
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.BigEndian, r.low)
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.BigEndian, r.high)
		if err != nil {
			return err
		}
	}
	return nil
}

func readProbabilityRanges(ranges *map[byte]ProbabilityRange, buf *bytes.Buffer) error {
	for {
		var char byte
		err := binary.Read(buf, binary.BigEndian, &char)
		if err != nil {
			break
		}
		var low, high float64
		err = binary.Read(buf, binary.BigEndian, &low)
		if err != nil {
			return err
		}
		err = binary.Read(buf, binary.BigEndian, &high)
		if err != nil {
			return err
		}
		(*ranges)[char] = ProbabilityRange{low, high}
	}
	return nil
}

func decodeArithmetic(data []byte, ranges map[byte]ProbabilityRange, length int) []byte {
	code := decodeBinary(data)
	decoded := make([]byte, 0, length)
	low, high := 0.0, 1.0

	for len(decoded) < length {
		r := (code - low) / (high - low)
		for char, pr := range ranges {
			if pr.low <= r && r < pr.high {
				decoded = append(decoded, char)
				rangeWidth := high - low
				high = low + rangeWidth*pr.high
				low = low + rangeWidth*pr.low
				break
			}
		}
	}
	return decoded
}

func decodeBinary(data []byte) float64 {
	code := 0.0
	for i := 0; i < len(data); i++ {
		code += float64(data[i]) * math.Pow(2, -float64(i+1))
	}
	return code
}
