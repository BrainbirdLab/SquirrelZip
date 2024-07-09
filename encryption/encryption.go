package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)
// Constants
const (
	MetadataLength = 1 // Length of the metadata
)

// Encrypt function
func Encrypt(data []byte, password string) ([]byte, error) {
	var metadata byte

	if password == "" {
		metadata = 0 // 0 indicates no password was used
		return append([]byte{metadata}, data...), nil
	}

	metadata = 1 // 1 indicates a password was used
	key, err := generateKey(password)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	encryptedData := gcm.Seal(nonce, nonce, data, nil)
	return append([]byte{metadata}, encryptedData...), nil
}

// Decrypt function
func Decrypt(data []byte, password string) ([]byte, error) {
	if len(data) < MetadataLength {
		return nil, errors.New("invalid data")
	}

	metadata := data[0]
	data = data[MetadataLength:]

	if metadata == 0 {
		// No password was used
		return data, nil
	}

	if password == "" {
		return nil, errors.New("password is required")
	}

	key, err := generateKey(password)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("incorrect password")
	}

	return plaintext, nil
}