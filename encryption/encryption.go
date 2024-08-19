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
	MetadataLength = 9 // Length of the metadata
)

// Encrypt function
func Encrypt(data *[]byte, password string) error {

	var metadata []byte

	if password == "" {
		metadata = append(metadata, 0)
	} else {
		metadata = append(metadata, 1)
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

	*data = append(metadata, *data...)

	return nil
}

// Decrypt function
func Decrypt(data *[]byte, password string) error {

	if len(*data) < MetadataLength {
		return fmt.Errorf("invalid data")
	}

	var hasPassword bool

	switch (*data)[0] {
	case 0:
		hasPassword = false
	case 1:
		hasPassword = true
	default:
		return fmt.Errorf("invalid metadata")
	}

	*data = (*data)[1:]

	if hasPassword {
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
	}

	return nil
}
