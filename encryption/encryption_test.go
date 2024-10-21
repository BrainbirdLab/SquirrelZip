package encryption

import (
	"bytes"
	"fmt"
	"testing"
)

var input []byte = []byte("Hello world")
var password string = "test1234"
var fatalEncrPassErr string = "failed to encrypt data with password: %v"
var fatalDecrPassErr string = "failed to decrypt data with password: %v"

const DECRYPT_SHOULD_FAIL = "should have failed"


func TestEncryptWithPassword(t *testing.T) {

	reader := bytes.NewReader(input)

	encryptedData := bytes.NewBuffer([]byte{})

	err := EncryptStream(reader, encryptedData, password)
	if err != nil {
		t.Fatalf(fatalEncrPassErr, err)
	}

	decryptedData := bytes.NewBuffer([]byte{})
	encryptedReader := bytes.NewReader(encryptedData.Bytes())

	err = DecryptStream(encryptedReader, decryptedData, password)
	if err != nil {
		t.Fatalf(fatalDecrPassErr, err)
	}

	if decryptedData.String() != string(input) {
		t.Fatalf("decrypted data does not match original data: decrypted(%v) != test(%v)", decryptedData.String(), input)
	}
}

func TestEncryptWithoutPassword(t *testing.T) {
	
	reader := bytes.NewReader(input)

	encryptedData := bytes.NewBuffer([]byte{})

	err := EncryptStream(reader, encryptedData, "")
	if err != nil {
		t.Fatalf(fatalEncrPassErr, err)
	}

	decryptedData := bytes.NewBuffer([]byte{})
	encryptedReader := bytes.NewReader(encryptedData.Bytes())

	err = DecryptStream(encryptedReader, decryptedData, "")
	if err != nil {
		t.Fatalf(fatalDecrPassErr, err)
	}

	if decryptedData.String() != string(input) {
		t.Fatalf("decrypted data does not match original data: decrypted(%v) != test(%v)", decryptedData.String(), input)
	}
}

func TestDecryptInvalidData(t *testing.T) {

	reader := bytes.NewReader(input)

	encryptedData := bytes.NewBuffer([]byte{})

	err := EncryptStream(reader, encryptedData, "")
	if err != nil {
		t.Fatalf(fatalEncrPassErr, err)
	}

	// write 3 in index 0 to make the data invalid
	encryptedData.Bytes()[0] = 3

	decryptedData := bytes.NewBuffer([]byte{})
	encryptedReader := bytes.NewReader(encryptedData.Bytes())

	err = DecryptStream(encryptedReader, decryptedData, "")
	if err == nil {
		t.Fatal(DECRYPT_SHOULD_FAIL)
	}
}

func TestDecryptInvalidPassword(t *testing.T) {
	
	reader := bytes.NewReader(input)

	encryptedData := bytes.NewBuffer([]byte{})

	err := EncryptStream(reader, encryptedData, password)
	if err != nil {
		t.Fatalf(fatalEncrPassErr, err)
	}

	decryptedData := bytes.NewBuffer([]byte{})
	encryptedReader := bytes.NewReader(encryptedData.Bytes())

	err = DecryptStream(encryptedReader, decryptedData, "invalid")
	if err == nil {
		t.Fatal(DECRYPT_SHOULD_FAIL)
	}

	fmt.Printf("Error successfully caught: %v\n", err)
}

func TestEncryptWithPassDecryptWithNoPass(t *testing.T) {
	
	reader := bytes.NewReader(input)

	encryptedData := bytes.NewBuffer([]byte{})

	err := EncryptStream(reader, encryptedData, password)
	if err != nil {
		t.Fatalf(fatalEncrPassErr, err)
	}

	decryptedData := bytes.NewBuffer([]byte{})
	encryptedReader := bytes.NewReader(encryptedData.Bytes())

	err = DecryptStream(encryptedReader, decryptedData, "")
	if err == nil {
		t.Fatal(DECRYPT_SHOULD_FAIL)
	}

	fmt.Printf("Error successfully caught: %v\n", err)
}