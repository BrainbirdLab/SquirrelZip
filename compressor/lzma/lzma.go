package lzma

import "fmt"

// Dictionary size for looking up previous occurrences
const (
	MaxWindowSize = 4096 // Size of sliding window
	MinMatchLen   = 3    // Minimum length for a match
)

type Match struct {
	offset int // Distance to the match
	length int // Length of the match
}

// compressData performs a simplified LZMA-style compression
func compressData(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return []byte{}, nil
	}

	var output []byte
	pos := 0
	dictionary := make(map[string]int)

	for pos < len(input) {
		// Find longest match in the sliding window
		match := findLongestMatch(input, pos, dictionary)

		if match.length >= MinMatchLen {
			// Encode as (offset, length) pair
			// Using simple format: [1][offset:2 bytes][length:1 byte]
			output = append(output, 1)
			output = append(output, byte(match.offset>>8), byte(match.offset))
			output = append(output, byte(match.length))

			// Update dictionary with all substrings in the match
			for i := 0; i < match.length; i++ {
				updateDictionary(input, pos+i, dictionary)
			}

			pos += match.length
		} else {
			// Encode literal byte
			// Format: [0][literal byte]
			output = append(output, 0)
			output = append(output, input[pos])

			updateDictionary(input, pos, dictionary)
			pos++
		}
	}

	return output, nil
}

// decompressData decodes the compressed data
func decompressData(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return []byte{}, nil
	}

	var output []byte
	pos := 0

	for pos < len(input) {
		flag := input[pos]
		pos++

		if flag == 0 {
			// Literal byte
			output = append(output, input[pos])
			pos++
		} else {
			// Match reference
			if pos+3 > len(input) {
				return nil, fmt.Errorf("invalid compressed data")
			}

			offset := int(input[pos])<<8 | int(input[pos+1])
			length := int(input[pos+2])
			pos += 3

			if offset > len(output) {
				return nil, fmt.Errorf("invalid offset")
			}

			// Copy matched bytes
			startPos := len(output) - offset
			for i := 0; i < length; i++ {
				output = append(output, output[startPos+i])
			}
		}
	}

	return output, nil
}

// findLongestMatch looks for the longest matching sequence in the window
func findLongestMatch(data []byte, pos int, dict map[string]int) Match {
	if pos >= len(data) {
		return Match{0, 0}
	}

	maxLen := MinMatchLen - 1
	bestMatch := Match{0, 0}

	// Try matching increasing lengths
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

// updateDictionary adds substrings to the dictionary
func updateDictionary(data []byte, pos int, dict map[string]int) {
	maxSubstringLen := MinMatchLen * 2
	for length := MinMatchLen; length <= maxSubstringLen && pos+length <= len(data); length++ {
		substr := string(data[pos : pos+length])
		dict[substr] = pos
	}
}
