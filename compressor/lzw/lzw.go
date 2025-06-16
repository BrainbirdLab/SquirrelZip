package lzw

import (
	"bufio"
	"fmt"
	"io"
)

const (
	maxDictSize     = 4096 // Maximum dictionary size (2^12)
	minDictSize     = 258  // Start of dynamic codes (0-255 are chars, 256 Clear, 257 EOI)
	startCodeSize   = 9    // Initial code size in bits
	maxCodeSize     = 12   // Maximum code size in bits
	clearCode       = 256  // LZW clear code
	endOfInfoCode   = 257  // LZW end of information code
)

// bitWriter helps write codes of varying bit lengths.
type bitWriter struct {
	w      io.Writer
	buffer uint64 // Buffer to accumulate bits
	nBits  uint   // Number of bits currently in the buffer
}

func newBitWriter(w io.Writer) *bitWriter {
	return &bitWriter{w: w}
}

// write writes a code with the given number of bits.
func (bw *bitWriter) write(code uint32, numBits uint) error {
	bw.buffer |= uint64(code) << bw.nBits
	bw.nBits += numBits
	for bw.nBits >= 8 {
		if _, err := bw.w.Write([]byte{byte(bw.buffer)}); err != nil {
			return err
		}
		bw.buffer >>= 8
		bw.nBits -= 8
	}
	return nil
}

// flush writes any remaining bits in the buffer.
func (bw *bitWriter) flush() error {
	if bw.nBits > 0 {
		if _, err := bw.w.Write([]byte{byte(bw.buffer)}); err != nil {
			return err
		}
		bw.buffer = 0
		bw.nBits = 0
	}
	return nil
}

// compressData compresses data from input reader and writes to output writer using LZW.
func compressData(input io.Reader, output io.Writer) error {
	dict := make(map[string]uint32)
	for i := 0; i < 256; i++ {
		dict[string(byte(i))] = uint32(i)
	}
	nextDictCodeVal := uint32(minDictSize)
	currentCodeSize := uint(startCodeSize)

	bw := newBitWriter(output)
	if err := bw.write(clearCode, currentCodeSize); err != nil {
		return fmt.Errorf("failed to write initial clear code: %w", err)
	}

	r := bufio.NewReader(input)
	var p []byte

	b, err := r.ReadByte()
	if err == io.EOF {
		if err := bw.write(endOfInfoCode, currentCodeSize); err != nil {
			return fmt.Errorf("failed to write EOI for empty input: %w", err)
		}
		return bw.flush()
	}
	if err != nil {
		return fmt.Errorf("failed to read first byte: %w", err)
	}
	p = append(p, b)

	for {
		b, err = r.ReadByte()
		if err == io.EOF {
			if codeP, okP := dict[string(p)]; okP {
				if errWrite := bw.write(codeP, currentCodeSize); errWrite != nil {
					return fmt.Errorf("failed to write last code for P: %w", errWrite)
				}
			} else {
				return fmt.Errorf("internal error: string P ('%s') not in dictionary at EOF", string(p))
			}
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read byte: %w", err)
		}

		c := []byte{b}
		pc := append(p, c...)

		if _, ok := dict[string(pc)]; ok {
			p = pc
		} else {
			if codeP, okP := dict[string(p)]; okP {
				if errWrite := bw.write(codeP, currentCodeSize); errWrite != nil {
					return fmt.Errorf("failed to write code for P: %w", errWrite)
				}
			} else {
				return fmt.Errorf("internal error: string P ('%s') not in dictionary", string(p))
			}

			if nextDictCodeVal < maxDictSize {
				dict[string(pc)] = nextDictCodeVal
				nextDictCodeVal++
				if nextDictCodeVal == (1<<currentCodeSize) && currentCodeSize < maxCodeSize {
					currentCodeSize++
				}
			}
			p = c
		}
	}

	if err := bw.write(endOfInfoCode, currentCodeSize); err != nil {
		return fmt.Errorf("failed to write EOI code: %w", err)
	}

	return bw.flush()
}

// bitReader helps read codes of varying bit lengths.
type bitReader struct {
	r      io.Reader
	buffer uint64
	nBits  uint
}

func newBitReader(r io.Reader) *bitReader {
	return &bitReader{r: r}
}

// read reads a code with the given number of bits.
func (br *bitReader) read(numBits uint) (uint32, error) {
	for br.nBits < numBits {
		var b [1]byte
		n, err := br.r.Read(b[:])
		if n == 0 && err == io.EOF {
			if br.nBits > 0 && br.nBits < numBits {
				return 0, io.ErrUnexpectedEOF
			}
			return 0, io.EOF
		}
		if err != nil && err != io.EOF {
			return 0, err
		}
		br.buffer |= uint64(b[0]) << br.nBits
		br.nBits += 8
		if err == io.EOF && br.nBits < numBits {
			return 0, io.ErrUnexpectedEOF
		}
	}

	code := uint32(br.buffer & ((1 << numBits) - 1))
	br.buffer >>= numBits
	br.nBits -= numBits
	return code, nil
}

// decompressData decompresses data from input reader and writes to output writer using LZW.
func decompressData(input io.Reader, output io.Writer) error {
	dict := make([][]byte, maxDictSize)
	for i := 0; i < 256; i++ {
		dict[i] = []byte{byte(i)}
	}
	nextCodeInTable := uint32(minDictSize)
	currentCodeSize := uint(startCodeSize)

	br := newBitReader(input)
	w := bufio.NewWriter(output)
	defer w.Flush()

	code, err := br.read(currentCodeSize)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return fmt.Errorf("stream too short, failed to read initial clear code: %w", err)
		}
		return fmt.Errorf("failed to read initial clear code: %w", err)
	}
	if code != clearCode {
		return fmt.Errorf("expected clearCode (%d) at start of stream, got %d", clearCode, code)
	}

	oldCode, err := br.read(currentCodeSize)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return fmt.Errorf("stream ended prematurely after clear code: %w", err)
		}
		return fmt.Errorf("failed to read first data code after clear code: %w", err)
	}

	if oldCode == endOfInfoCode {
		return nil
	}
	if oldCode > 255 {
		return fmt.Errorf("first data code %d after clear must be a literal (<256)", oldCode)
	}

	s := dict[oldCode]
	if _, err := w.Write(s); err != nil {
		return fmt.Errorf("failed to write first string: %w", err)
	}

	var prevEntry []byte
	prevEntry = append(prevEntry[:0], s...)

	for {
		newCode, err := br.read(currentCodeSize)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// This is not necessarily an error if we just finished reading the last code.
				// The loop condition newCode == endOfInfoCode should handle graceful exit.
				// However, if newCode is not EOI, then it's an unexpected EOF.
				// We let the EOI check below handle it. If it's not EOI, then it's an error.
			} else {
				return fmt.Errorf("error reading new code: %w", err)
			}
		}

		if newCode == endOfInfoCode {
            if err == io.EOF && br.nBits > 0 { // Check if EOI is partial
                 return fmt.Errorf("stream ended prematurely with partial EOI code: %w", io.ErrUnexpectedEOF)
            }
			break
		}
        // If we hit EOF above but newCode wasn't EOI, it means stream ended before EOI.
        if err == io.EOF && newCode != endOfInfoCode {
            return fmt.Errorf("stream ended before EOI code: %w", io.ErrUnexpectedEOF)
        }
         if err == io.ErrUnexpectedEOF { // Propagate if specifically this error
            return fmt.Errorf("unexpected EOF while reading code: %w", err)
        }


		if newCode == clearCode {
			currentCodeSize = startCodeSize
			nextCodeInTable = minDictSize
			// Reset dictionary table entries from minDictSize upwards
            // The first 256 entries (literals) remain unchanged.
            // No explicit clearing needed for `dict` slice entries beyond `nextCodeInTable`
            // as they will be overwritten or not accessed.

			oldCode, err = br.read(currentCodeSize)
			if err != nil {
                if err == io.EOF || err == io.ErrUnexpectedEOF {
                    return fmt.Errorf("stream ended prematurely after clear code during reset: %w", err)
                }
				return fmt.Errorf("failed to read code after clear code: %w", err)
			}
			if oldCode == endOfInfoCode {
				break
			}
			if oldCode >= minDictSize { // After clear, next code must be a literal or EOI
				return fmt.Errorf("expected literal code (<256) after clear code, got %d", oldCode)
			}
			s = dict[oldCode]
			if _, err := w.Write(s); err != nil {
				return fmt.Errorf("failed to write string after clear code: %w", err)
			}
			prevEntry = append(prevEntry[:0], s...)
			continue
		}

		var currentEntry []byte
		if newCode < nextCodeInTable {
			currentEntry = dict[newCode]
		} else if newCode == nextCodeInTable {
			if len(prevEntry) == 0 {
				return fmt.Errorf("internal error: prevEntry is empty in KWKWK case for code %d", newCode)
			}
			currentEntry = append(prevEntry, prevEntry[0])
		} else {
			return fmt.Errorf("invalid code %d encountered (nextCodeInTable: %d, currentCodeSize: %d bits, prevCode: %d)", newCode, nextCodeInTable, currentCodeSize, oldCode)
		}

		if _, err := w.Write(currentEntry); err != nil {
			return fmt.Errorf("failed to write current entry: %w", err)
		}

		if nextCodeInTable < maxDictSize {
			if len(prevEntry) == 0 {
				return fmt.Errorf("internal error: prevEntry is empty when forming new dict entry for code %d", nextCodeInTable)
			}
            if len(currentEntry) == 0 {
                 return fmt.Errorf("internal error: currentEntry is empty when forming new dict entry for code %d, newCode %d", nextCodeInTable, newCode)
            }
			newDictString := append(prevEntry, currentEntry[0])
			dict[nextCodeInTable] = newDictString
			nextCodeInTable++

			if nextCodeInTable == (1<<currentCodeSize) && currentCodeSize < maxCodeSize {
				currentCodeSize++
			}
		}
		prevEntry = append(prevEntry[:0], currentEntry...)
		oldCode = newCode // Update oldCode for the next iteration
	}
	return w.Flush()
}
