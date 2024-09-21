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
	NO_PASSWORD byte = 43
	PASSWORD    byte = 57
	BufferSize     = 1024
)

// EncryptStream encrypts data from an io.Reader and writes the encrypted data to an io.Writer
func EncryptStream(reader io.Reader, writer io.Writer, password string) error {
	// Write the metadata based on whether a password is provided
	if err := writeMetadata(writer, password); err != nil {
		return err
	}

	// Encrypt with or without a password
	if password == "" {
		return copyData(reader, writer)
	}
	return encryptWithPassword(reader, writer, password)
}

// DecryptStream decrypts data from an io.Reader and writes the decrypted data to an io.Writer
func DecryptStream(reader io.Reader, writer io.Writer, password string) error {
	// Parse metadata to determine if password is required
	hasPassword, err := readMetadata(reader)
	if err != nil {
		return err
	}

	// If no password was used, copy the data directly
	if !hasPassword {
		return copyData(reader, writer)
	}

	// Decrypt the data
	return decryptWithPassword(reader, writer, password)
}

// writeMetadata writes metadata indicating if password encryption is being used
func writeMetadata(writer io.Writer, password string) error {
	var metadata []byte
	if password == "" {
		metadata = append(metadata, NO_PASSWORD) // No password
	} else {
		metadata = append(metadata, PASSWORD) // Password used
	}
	_, err := writer.Write(metadata)
	return err
}

// readMetadata reads metadata and determines if password encryption was used
func readMetadata(reader io.Reader) (bool, error) {
	metadata := make([]byte, 1)
	if _, err := io.ReadFull(reader, metadata); err != nil {
		return false, fmt.Errorf("failed to read metadata: %v", err)
	}
	switch metadata[0] {
	case NO_PASSWORD:
		return false, nil
	case PASSWORD:
		return true, nil
	default:
		return false, fmt.Errorf("invalid metadata")
	}
}

// copyData simply copies data from the reader to the writer without encryption
func copyData(reader io.Reader, writer io.Writer) error {
	_, err := io.Copy(writer, reader)
	return err
}

// encryptWithPassword encrypts data using AES-GCM with a password
func encryptWithPassword(reader io.Reader, writer io.Writer, password string) error {
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

	nonce, err := generateNonce(gcm)
	if err != nil {
		return err
	}

	// Write the nonce to the writer
	if _, err := writer.Write(nonce); err != nil {
		return err
	}

	// Encrypt and write the data in chunks
	return processStream(reader, writer, gcm, nonce)
}

// decryptWithPassword decrypts data using AES-GCM with a password
func decryptWithPassword(reader io.Reader, writer io.Writer, password string) error {
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

	// Read and validate the nonce
	nonceSize := gcm.NonceSize()
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(reader, nonce); err != nil {
		return fmt.Errorf("failed to read nonce: %v", err)
	}

	// Decrypt and write the data in chunks
	return decryptStream(reader, writer, gcm, nonce)
}

// generateNonce creates a random nonce
func generateNonce(gcm cipher.AEAD) ([]byte, error) {
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}
	return nonce, nil
}

// processStream reads from the reader, encrypts each chunk, and writes it to the writer
func processStream(reader io.Reader, writer io.Writer, gcm cipher.AEAD, nonce []byte) error {
	buf := make([]byte, BufferSize)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		// Encrypt the chunk and write it
		ciphertext := gcm.Seal(nil, nonce, buf[:n], nil)
		if _, err := writer.Write(ciphertext); err != nil {
			return err
		}
	}
	return nil
}

// decryptStream reads from the reader, decrypts each chunk, and writes it to the writer
func decryptStream(reader io.Reader, writer io.Writer, gcm cipher.AEAD, nonce []byte) error {
	buf := make([]byte, BufferSize+gcm.Overhead())
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		// Decrypt the chunk and write it
		plaintext, err := gcm.Open(nil, nonce, buf[:n], nil)
		if err != nil {
			return fmt.Errorf("decryption failed: %v", err)
		}
		if _, err := writer.Write(plaintext); err != nil {
			return err
		}
	}
	return nil
}
