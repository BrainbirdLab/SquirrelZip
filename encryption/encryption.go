package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"file-compressor/constants"
)


// EncryptStream reads data from the provided reader, encrypts it, and writes the encrypted data to the provided writer.
// If a password is provided, the data will be encrypted using the password. If no password is provided, the data will
// be copied without encryption.
//
// Parameters:
//   - reader: An io.Reader from which the data will be read.
//   - writer: An io.Writer to which the encrypted data will be written.
//   - password: A string used as the password for encryption. If empty, no encryption will be applied.
//
// Returns:
//   - error: An error if any occurs during the encryption or writing process, otherwise nil.
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

// DecryptStream reads encrypted data from the provided reader, decrypts it using the given password,
// and writes the decrypted data to the provided writer. If the data does not require a password,
// it is copied directly from the reader to the writer.
//
// Parameters:
//   - reader: An io.Reader from which the encrypted data is read.
//   - writer: An io.Writer to which the decrypted data is written.
//   - password: A string containing the password used for decryption.
//
// Returns:
//   - error: An error if any issues occur during the decryption process, or nil if successful.
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


// writeMetadata writes metadata to the provided writer indicating whether a password is used.
// If the password is an empty string, it writes a constant indicating no password is used.
// Otherwise, it writes a constant indicating a password is used.
//
// Parameters:
//   - writer: An io.Writer where the metadata will be written.
//   - password: A string representing the password. If empty, it indicates no password.
//
// Returns:
//   - error: An error if writing to the writer fails, otherwise nil.
func writeMetadata(writer io.Writer, password string) error {
	var metadata []byte
	if password == "" {
		metadata = append(metadata, constants.NO_PASSWORD) // No password
	} else {
		metadata = append(metadata, constants.PASSWORD) // Password used
	}
	_, err := writer.Write(metadata)
	return err
}

// readMetadata reads a single byte of metadata from the provided io.Reader.
// It returns a boolean indicating whether a password is required and an error if any occurs during reading.
// The metadata byte is interpreted as follows:
// - constants.NO_PASSWORD: returns false, nil
// - constants.PASSWORD: returns true, nil
// - Any other value: returns false, fmt.Errorf("invalid metadata")
//
// Parameters:
// - reader: an io.Reader from which the metadata byte is read.
//
// Returns:
// - bool: true if a password is required, false otherwise.
// - error: an error if there is an issue reading the metadata or if the metadata is invalid.
func readMetadata(reader io.Reader) (bool, error) {
	metadata := make([]byte, 1)
	if _, err := io.ReadFull(reader, metadata); err != nil {
		return false, fmt.Errorf("failed to read metadata: %v", err)
	}
	switch metadata[0] {
	case constants.NO_PASSWORD:
		return false, nil
	case constants.PASSWORD:
		return true, nil
	default:
		return false, fmt.Errorf("invalid metadata")
	}
}

// simply copies data from the reader to the writer without encryption
func copyData(reader io.Reader, writer io.Writer) error {
	_, err := io.Copy(writer, reader)
	return err
}


// encryptWithPassword encrypts data from the provided reader using the given password
// and writes the encrypted data to the provided writer.
//
// Parameters:
//   - reader: An io.Reader from which the plaintext data is read.
//   - writer: An io.Writer to which the encrypted data is written.
//   - password: A string used to generate the encryption key.
//
// Returns:
//   - error: An error if any step of the encryption process fails, otherwise nil.
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


// decryptWithPassword decrypts data from the provided reader using the given password
// and writes the decrypted data to the provided writer. It returns an error if the
// decryption process fails at any step.
//
// Parameters:
// - reader: an io.Reader from which the encrypted data is read.
// - writer: an io.Writer to which the decrypted data is written.
// - password: a string used to derive the decryption key.
//
// Returns:
// - error: an error if the decryption fails, or nil if the decryption is successful.
//
// The function performs the following steps:
// 1. Validates that the password is not empty.
// 2. Generates a decryption key from the password.
// 3. Creates a new AES cipher block using the generated key.
// 4. Creates a Galois/Counter Mode (GCM) cipher from the AES block.
// 5. Reads and validates the nonce from the reader.
// 6. Decrypts the data in chunks and writes it to the writer.
func decryptWithPassword(reader io.Reader, writer io.Writer, password string) error {

	if password == "" {
		return fmt.Errorf("password required for decryption")
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

	// Read and validate the nonce
	nonceSize := gcm.NonceSize()
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(reader, nonce); err != nil {
		return fmt.Errorf("failed to read nonce: %v", err)
	}

	// Decrypt and write the data in chunks
	return decryptStream(reader, writer, gcm, nonce)
}


// generateNonce generates a nonce of the appropriate size for the given
// AEAD cipher. It uses a cryptographically secure random number generator
// to fill the nonce with random bytes.
//
// Parameters:
// - gcm: An AEAD cipher instance which provides the nonce size.
//
// Returns:
// - A byte slice containing the generated nonce.
// - An error if the nonce generation fails.
func generateNonce(gcm cipher.AEAD) ([]byte, error) {
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}
	return nonce, nil
}


// processStream reads data from the provided io.Reader, encrypts it using the given
// cipher.AEAD and nonce, and writes the encrypted data to the provided io.Writer.
// 
// Parameters:
//   - reader: an io.Reader from which the data is read.
//   - writer: an io.Writer to which the encrypted data is written.
//   - gcm: a cipher.AEAD instance used for encryption.
//   - nonce: a byte slice used as the nonce for encryption.
//
// Returns:
//   - error: an error if any occurs during reading, encrypting, or writing the data.
//
// The function reads data in chunks of size constants.BUFFER_SIZE, encrypts each chunk
// using the provided AEAD cipher and nonce, and writes the encrypted chunk to the writer.
// The process continues until the end of the reader is reached.
func processStream(reader io.Reader, writer io.Writer, gcm cipher.AEAD, nonce []byte) error {
	buf := make([]byte, constants.BUFFER_SIZE)
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


// decryptStream decrypts data from the provided io.Reader and writes the decrypted data to the provided io.Writer.
// It uses the given cipher.AEAD and nonce for decryption.
//
// Parameters:
//   - reader: an io.Reader from which encrypted data is read.
//   - writer: an io.Writer to which decrypted data is written.
//   - gcm: a cipher.AEAD instance used for decryption.
//   - nonce: a byte slice containing the nonce used for decryption.
//
// Returns:
//   - error: an error if decryption or writing fails, otherwise nil.
func decryptStream(reader io.Reader, writer io.Writer, gcm cipher.AEAD, nonce []byte) error {
	buf := make([]byte, constants.BUFFER_SIZE+gcm.Overhead())
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
