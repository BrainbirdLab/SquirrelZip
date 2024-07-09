package encryption

import (
	"errors"
)

// Function to generate key from password
func generateKey(password string) ([]byte, error) {
	key := []byte(password)
	if len(key) > 32 {
		return nil, errors.New("password too long")
	}
	for len(key) < 32 {
		key = append(key, '0')
	}
	return key, nil
}