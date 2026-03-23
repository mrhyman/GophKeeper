package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	keyLen     = 32
	saltLen    = 16
	iterations = 100_000
)

func DeriveKey(masterPassword string, salt []byte) []byte {
	return pbkdf2.Key([]byte(masterPassword), salt, iterations, keyLen, sha256.New)
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func Encrypt(plaintext []byte, masterPassword string) ([]byte, error) {
	salt, err := GenerateSalt()
	if err != nil {
		return nil, err
	}

	key := DeriveKey(masterPassword, salt)

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

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// salt + nonce + ciphertext
	result := make([]byte, 0, saltLen+len(nonce)+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

func Decrypt(encrypted []byte, masterPassword string) ([]byte, error) {
	if len(encrypted) < saltLen+12 {
		return nil, errors.New("encrypted data too short")
	}

	salt := encrypted[:saltLen]
	key := DeriveKey(masterPassword, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < saltLen+nonceSize {
		return nil, errors.New("encrypted data too short")
	}

	nonce := encrypted[saltLen : saltLen+nonceSize]
	ciphertext := encrypted[saltLen+nonceSize:]

	return gcm.Open(nil, nonce, ciphertext, nil)
}
