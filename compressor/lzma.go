package compressor

import (
	"bytes"
	"fmt"
)

type literalEncoder struct {
	probs [256]uint16
}

func makeLiteralEncoder() literalEncoder {
	var le literalEncoder
	for i := range le.probs {
		le.probs[i] = 1024
	}
	return le
}

func (le *literalEncoder) updateSymbol(symbol byte) {
	le.probs[symbol] += (2048 - le.probs[symbol]) >> 5
}

func (le *literalEncoder) getSymbolProbs(symbol byte) uint16 {
	return le.probs[symbol]
}

func (le *literalEncoder) updateLiteralProbabilities() {
	for i := range le.probs {
		le.probs[i] -= le.probs[i] >> 5
	}
}

func (le *literalEncoder) getLiteralProbs() [256]uint16 {
	return le.probs
}

func newRangeEncoder(w *bytes.Buffer) *rangeEncoder {
	return &rangeEncoder{
		writer: w,
		low:    0,
		high:   0xFFFFFFFF,
	}
}

type rangeEncoder struct {
	writer *bytes.Buffer
	low    uint32
	high   uint32
	cache  uint32
	cached uint
}

func (re *rangeEncoder) init() {
	re.low = 0
	re.high = 0xFFFFFFFF
	re.cache = 0
	re.cached = 0
}

func (re *rangeEncoder) finish() {
	for i := 0; i < 5; i++ {
		re.shiftLow()
	}
}

func (re *rangeEncoder) shiftLow() {
	if re.low>>24 != re.high>>24 {
		if re.high>>23 == 1 {
			re.low <<= 1
			re.high = (re.high << 1) | 1
			re.high ^= 0x80000000
			re.low ^= 0x80000000
		} else {
			re.low <<= 1
			re.high = (re.high << 1) | 1
		}
	}

	re.low <<= 1
	re.high = ((re.high << 1) | 1) & 0xFFFFFFFF
}

func (re *rangeEncoder) encode(le, le2 *literalEncoder, symbol byte) {
	probs := le.getLiteralProbs()
	probs2 := le2.getLiteralProbs()
	prob := probs[symbol]
	prob2 := probs2[symbol]

	totalProb := prob + prob2
	if totalProb > 0 {
		le.updateSymbol(symbol)
		le2.updateSymbol(symbol)
	} else {
		totalProb = 1
	}

	le.updateLiteralProbabilities()
	le2.updateLiteralProbabilities()

	// Update range
	rangeSize := re.high - re.low + 1
	re.high = re.low + (rangeSize*uint32(totalProb)/2048 - 1)
	re.low = re.low + (rangeSize*uint32(prob)/2048)

	// Normalize
	for {
		if re.high < 0x80000000 {
			re.shiftLow()
		} else if re.low >= 0x80000000 {
			re.low -= 0x80000000
			re.high -= 0x80000000
		} else if re.low >= 0x40000000 && re.high < 0xC0000000 {
			re.cache++
			re.low -= 0x40000000
			re.high -= 0x40000000
		} else {
			break
		}

		re.low <<= 1
		re.high = (re.high << 1) | 1
		re.high &= 0xFFFFFFFF
		re.low &= 0xFFFFFFFF
	}

	// Write bits
	for {
		if re.high < 0x80000000 {
			re.writer.WriteByte(byte(re.cache))
			for i := uint(0); i < re.cached; i++ {
				re.writer.WriteByte(byte(re.cache >> 8))
				re.cache >>= 8
			}
			re.cached = 0
			re.cache = 0xFF
		} else if re.low >= 0x80000000 {
			re.writer.WriteByte(byte(re.cache + 1))
			for i := uint(0); i < re.cached; i++ {
				re.writer.WriteByte(byte(re.cache))
				re.cache >>= 8
			}
			re.cached = 0
			re.cache = 0
		} else if re.low >= 0x40000000 && re.high < 0xC0000000 {
			re.cached++
			re.low -= 0x40000000
			re.high -= 0x40000000
		} else {
			break
		}

		re.low <<= 1
		re.high = (re.high << 1) | 1
		re.high &= 0xFFFFFFFF
		re.low &= 0xFFFFFFFF
	}
}

func newRangeDecoder(r *bytes.Reader) *rangeDecoder {
	return &rangeDecoder{
		reader: r,
		low:    0,
		high:   0xFFFFFFFF,
		code:   0,
	}
}

type rangeDecoder struct {
	reader *bytes.Reader
	low    uint32
	high   uint32
	code   uint32
}

func (rd *rangeDecoder) init() error {
	rd.low = 0
	rd.high = 0xFFFFFFFF
	rd.code = 0
	
	for i := 0; i < 5; i++ {
		byt, err := rd.reader.ReadByte()
		if err != nil {
			return err
		}
		rd.code = (rd.code << 8) | uint32(byt)
	}

	return nil
}

func (rd *rangeDecoder) decodeLiteral(le, le2 *literalEncoder) (byte, error) {
	probs := le.getLiteralProbs()
	probs2 := le2.getLiteralProbs()

	totalProb := uint32(probs[0] + probs2[0])
	if totalProb == 0 {
		totalProb = 1
	}

	// Normalize
	rangeSize := rd.high - rd.low + 1
	value := ((rd.code - rd.low + 1) * 2048 - 1) / rangeSize
	symbol := byte(0)
	for i := byte(0); i < 255; i++ {
		if value < totalProb {
			symbol = i
			break
		}
		value -= totalProb
		totalProb = uint32(probs[i+1] + probs2[i+1])
		if totalProb == 0 {
			totalProb = 1
		}
	}

	// Update range
	prob := probs[symbol]
	prob2 := probs2[symbol]
	le.updateSymbol(symbol)
	le2.updateSymbol(symbol)
	le.updateLiteralProbabilities()
	le2.updateLiteralProbabilities()
	rd.high = rd.low + (rangeSize*uint32(prob+prob2)/2048 - 1)
	rd.low = rd.low + (rangeSize*uint32(prob)/2048)

	// Normalize
	for {
		if rd.high < 0x80000000 {
			rd.low <<= 1
			rd.high = (rd.high << 1) | 1
			rd.high &= 0xFFFFFFFF
			byt, err := rd.reader.ReadByte()
			if err != nil {
				return 0, err
			}
			rd.code = (rd.code << 1) | uint32(byt)
		} else if rd.low >= 0x80000000 {
			rd.low -= 0x80000000
			rd.high -= 0x80000000
			rd.code -= 0x80000000
		} else if rd.low >= 0x40000000 && rd.high < 0xC0000000 {
			rd.low -= 0x40000000
			rd.high -= 0x40000000
			rd.code -= 0x40000000
		} else {
			break
		}

		rd.low <<= 1
		rd.high = (rd.high << 1) | 1
		rd.high &= 0xFFFFFFFF
		rd.low &= 0xFFFFFFFF
		byt, err := rd.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		rd.code = (rd.code << 1) | uint32(byt)
	}

	return symbol, nil
}

// LZMA represents the LZMA algorithm.
type LZMA struct {
	literalEncoder  literalEncoder
	literalEncoder2 literalEncoder
}

// NewLZMA creates a new instance of LZMA compressor.
func NewLZMA() *LZMA {
	return &LZMA{
		literalEncoder:  makeLiteralEncoder(),
		literalEncoder2: makeLiteralEncoder(),
	}
}

// Compress compresses the input data using LZMA algorithm.
func (lzma *LZMA) Compress(data []byte) []byte {
	var buf bytes.Buffer
	re := newRangeEncoder(&buf)
	re.init()

	for _, b := range data {
		re.encode(&lzma.literalEncoder, &lzma.literalEncoder2, b)
	}

	re.finish()
	return buf.Bytes()
}

// Decompress decompresses the input data using LZMA algorithm.
func (lzma *LZMA) Decompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	rd := newRangeDecoder(bytes.NewReader(data))
	rd.init()

	for i := 0; i < len(data)-5; i++ {
		byt, err := rd.decodeLiteral(&lzma.literalEncoder, &lzma.literalEncoder2)
		if err != nil {
			return nil, err
		}
		buf.Write([]byte{byt})
	}

	return buf.Bytes(), nil
}

func compress(input []byte) []byte {
	return NewLZMA().Compress(input)
}

func decompress(input []byte) []byte {
	data, err := NewLZMA().Decompress(input)
	if err != nil {
		panic(err)
	}

	return data
}


// Helper function to test LZMA compression and decompression.
func test() {
	input := []byte("hello world")
	output := compress(input)
	fmt.Println("Compressed data: ", output)

	decompressed := decompress(output)
	fmt.Println("Decompressed data: ", string(decompressed))
}
