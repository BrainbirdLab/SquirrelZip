package lz77

import (
    "bytes"
    "encoding/binary"
    "io"
)

const (
    WindowSize    = 4096 // Search window size
    BufferSize    = 16   // Look-ahead buffer size
    MinMatchLen   = 3    // Minimum match length
)

type Match struct {
    Distance uint16 // Distance to the match in the window
    Length   uint8  // Length of the match
    NextByte byte   // Next byte after match
}

func findLongestMatch(data []byte, currentPos int) Match {
    windowStart := max(0, currentPos-WindowSize)
    lookAheadEnd := min(len(data), currentPos+BufferSize)
    
    if currentPos >= len(data) {
        return Match{0, 0, 0}
    }

    bestLength := uint8(0)
    bestDistance := uint16(0)
    nextByte := data[currentPos]

    if lookAheadEnd-currentPos >= MinMatchLen {
        for i := windowStart; i < currentPos; i++ {
            matchLength := 0
            for currentPos+matchLength < lookAheadEnd && 
                i+matchLength < currentPos &&
                data[i+matchLength] == data[currentPos+matchLength] {
                matchLength++
            }

            if matchLength >= MinMatchLen && uint8(matchLength) > bestLength {
                bestLength = uint8(matchLength)
                bestDistance = uint16(currentPos - i)
                if currentPos+matchLength < len(data) {
                    nextByte = data[currentPos+matchLength]
                }
            }
        }
    }

    return Match{bestDistance, bestLength, nextByte}
}

func Compress(input []byte) ([]byte, error) {
    var output bytes.Buffer
    pos := 0
    inputLen := len(input)

    for pos < inputLen {
        match := findLongestMatch(input, pos)
        
        if err := binary.Write(&output, binary.LittleEndian, match.Distance); err != nil {
            return nil, err
        }
        if err := binary.Write(&output, binary.LittleEndian, match.Length); err != nil {
            return nil, err
        }
        if err := binary.Write(&output, binary.LittleEndian, match.NextByte); err != nil {
            return nil, err
        }

        if match.Length > 0 {
            pos += int(match.Length)
        }
        pos++
    }

    return output.Bytes(), nil
}

func Decompress(input []byte) ([]byte, error) {
    var output bytes.Buffer
    reader := bytes.NewReader(input)

    for {
        var match Match
        
        err := binary.Read(reader, binary.LittleEndian, &match.Distance)
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }

        err = binary.Read(reader, binary.LittleEndian, &match.Length)
        if err != nil {
            return nil, err
        }

        err = binary.Read(reader, binary.LittleEndian, &match.NextByte)
        if err != nil {
            return nil, err
        }

        if match.Length > 0 {
            start := output.Len() - int(match.Distance)
            for i := 0; i < int(match.Length); i++ {
                if start+i >= 0 && start+i < output.Len() {
                    b := output.Bytes()[start+i]
                    output.WriteByte(b)
                }
            }
        }
        if pos := output.Len(); pos < len(input) {
            output.WriteByte(match.NextByte)
        }
    }

    return output.Bytes(), nil
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}