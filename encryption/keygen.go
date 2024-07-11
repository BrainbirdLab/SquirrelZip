package encryption

import "fmt"

// Function to generate key from password
func generateKey(password string) ([]byte, error) {
	key := []byte(password)
	if len(key) > 32 {
		return nil, fmt.Errorf("password too long")
	}
	for len(key) < 32 {
		key = append(key, '0')
	}
	return key, nil
}
