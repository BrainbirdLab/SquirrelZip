package encryption

import (
	"testing"
)

func TestEncryptWithPassword(t *testing.T) {
	input := []byte("Hello world")
	password := "test1234"

	output := input

	err := Encrypt(&output, password)
	if err != nil {
		t.Fatalf("failed to encrypt data with password: %v", err)
	}

	err = Decrypt(&output, password)

	if err != nil {
		t.Fatalf("failed to decrypt data with password: %v", err)
	}

	if string(output) != string(input) {
		t.Fatalf("decrypted data does not match original data: decrypted(%v) != test(%v)", string(output), string(input))
	}
}

func TestEncryptWithoutPassword(t *testing.T) {
	input := []byte("Hello world")
	password := ""

	output := input

	err := Encrypt(&output, password)
	if err != nil {
		t.Fatalf("failed to encrypt data without password: %v", err)
	}

	err = Decrypt(&output, password)

	if err != nil {
		t.Fatalf("failed to decrypt data without password: %v", err)
	}

	if string(output) != string(input) {
		t.Fatalf("decrypted data does not match original data: decrypted(%v) != test(%v)", string(output), input)
	}
}

func TestDecryptInvalidData(t *testing.T) {

	input := []byte("Hello world")
	password := "test1234"

	output := input

	err := Encrypt(&output, password)
	if err != nil {
		t.Fatalf("failed to encrypt data with password: %v", err)
	}

	// Modify the encrypted data
	output[0] = 2

	err = Decrypt(&output, password)

	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}

func TestDecryptInvalidPassword(t *testing.T) {
	input := []byte("Hello world")
	password := "test1234"

	output := input

	err := Encrypt(&output, password)
	if err != nil {
		t.Fatalf("failed to encrypt data with password: %v", err)
	}

	invalidPass := "invalid password"

	err = Decrypt(&output, invalidPass)

	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}