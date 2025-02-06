package lzma

import (
	"bytes"
	"testing"
)

func TestCompressDecompress(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Empty input", "", false},
		{"Short input", "abc", false},
		{"Long input", "The quick brown fox jumps over the lazy dog", false},
		{"Repeating input", "aaaaaa", false},
		{"Very long input", "this is a simple example of lzma compression algorithm. this is a simple example of lzma compression algorithm. Let me tell you a story. Once upon a time, there was a little duckling named Ugly. He was very sad because he was different from his siblings. They were all beautiful, but he was ugly. He was so ugly that everyone called him Ugly Duckling. He was so sad that he ran away from home. He wandered around the countryside, looking for a place where he could be happy. He met many animals, but none of them wanted to be his friend. He was so lonely that he cried himself to sleep every night. One day, he met a kind old man who took him in and cared for him. Soon he was happy again. He grew up to be a beautiful swan. He was no longer an ugly duckling. He was a beautiful swan. The end.", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var compressed bytes.Buffer
			_, err := compressData(bytes.NewReader([]byte(tt.input)), &compressed)
			if (err != nil) != tt.wantErr {
				t.Errorf("compressData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var decompressed bytes.Buffer
			err = decompressData(&compressed, &decompressed, uint64(compressed.Len()))
			if (err != nil) != tt.wantErr {
				t.Errorf("decompressData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if decompressed.String() != tt.input {
				t.Errorf("decompressed data = %v, want %v", decompressed.String(), tt.input)
			}
		})
	}
}
