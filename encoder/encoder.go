package encoder

import (
	"bytes"
	"encoding/binary"
	"file-compressor/utils"
)

func ParseDecompressedContent(files *[]utils.File, numFiles *uint32, decompressedContentBuf *bytes.Buffer) error {
	for f := uint32(0); f < *numFiles; f++ {
		// Read filename length
		var nameLength uint32
		err := binary.Read(decompressedContentBuf, binary.BigEndian, &nameLength)
		if err != nil {
			return err
		}
		// Read filename
		name := make([]byte, nameLength)
		_, err = decompressedContentBuf.Read(name)
		if err != nil {
			return err
		}
		// Read content length
		var contentLength uint32
		err = binary.Read(decompressedContentBuf, binary.BigEndian, &contentLength)
		if err != nil {
			return err
		}
		// Read content
		content := make([]byte, contentLength)
		_, err = decompressedContentBuf.Read(content)
		if err != nil {
			return err
		}
		*files = append(*files, utils.File{
			Name:    string(name),
			Content: content,
		})
	}
	return nil
}

func CreateContentBuffer(files []utils.File, rawContent *bytes.Buffer) error {
	for _, file := range files {
		// Write filename length and filename
		filenameLen := uint32(len(file.Name))
		err := binary.Write(rawContent, binary.BigEndian, filenameLen)
		if err != nil {
			return err
		}
		_, err = rawContent.WriteString(file.Name)
		if err != nil {
			return err
		}

		// Write content length and content
		contentLen := uint32(len(file.Content))
		err = binary.Write(rawContent, binary.BigEndian, contentLen)
		if err != nil {
			return err
		}
		_, err = rawContent.Write(file.Content)
		if err != nil {
			return err
		}
	}
	return nil
}