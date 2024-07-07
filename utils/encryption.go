package utils

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/md5"
    "crypto/rand"
    "errors"
    "io"
)

// generateKey creates a key of the appropriate length for AES encryption.
func generateKey(password string) ([]byte, error) {
    if len(password) == 0 {
        return nil, errors.New("password cannot be empty")
    }
    hash := md5.New()
    hash.Write([]byte(password))
    return hash.Sum(nil), nil
}

func Encrypt(data []byte, password string) []byte {
    key, err := generateKey(password)
    if err != nil {
        panic(err)
    }
    block, err := aes.NewCipher(key)
    if err != nil {
        panic(err)
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        panic(err)
    }
    nonce := make([]byte, gcm.NonceSize())
    io.ReadFull(rand.Reader, nonce)
    return gcm.Seal(nonce, nonce, data, nil)
}

func Decrypt(data []byte, password string) ([]byte, error) {
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
    return gcm.Open(nil, nonce, ciphertext, nil)
}
