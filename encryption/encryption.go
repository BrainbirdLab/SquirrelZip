package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Constants
const (
	MetadataLength = 1 // Length of the metadata
)

// Encrypt function
func Encrypt(data *[]byte, password string) error {
	var metadata byte

	if password == "" {
		metadata = 0 // 0 indicates no password was used
	} else {
		metadata = 1 // 1 indicates a password was used
		key, err := generateKey(password)
		if err != nil {
			return err
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return err
		}
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return err
		}
		*data = gcm.Seal(nonce, nonce, *data, nil)
	}
	*data = append([]byte{metadata}, *data...)
	return nil
}

// Decrypt function
func Decrypt(data *[]byte, password string) error {

	if len(*data) < MetadataLength {
		return fmt.Errorf("invalid data")
	}

	metadata := (*data)[0]
	*data = (*data)[MetadataLength:]
	if metadata == 0 {
		return nil
	} else if metadata != 1 {
		return fmt.Errorf("invalid metadata")
	}

	if password == "" {
		return fmt.Errorf("password is required")
	}

	key, err := generateKey(password)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonceSize := gcm.NonceSize()
	if len(*data) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := (*data)[:nonceSize], (*data)[nonceSize:]

	*data, err = gcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		return fmt.Errorf("incorrect password")
	}

	return nil
}
