package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	require.NoError(t, err)
	assert.Len(t, salt1, saltLen)

	salt2, err := GenerateSalt()
	require.NoError(t, err)
	assert.Len(t, salt2, saltLen)

	// Соли должны быть разными
	assert.False(t, bytes.Equal(salt1, salt2))
}

func TestDeriveKey(t *testing.T) {
	salt := []byte("1234567890123456")
	password := "mypassword"

	key1 := DeriveKey(password, salt)
	assert.Len(t, key1, keyLen)

	// Одинаковые входные данные — одинаковый ключ
	key2 := DeriveKey(password, salt)
	assert.True(t, bytes.Equal(key1, key2))

	// Разные пароли — разные ключи
	key3 := DeriveKey("differentpassword", salt)
	assert.False(t, bytes.Equal(key1, key3))

	// Разные соли — разные ключи
	key4 := DeriveKey(password, []byte("6543210987654321"))
	assert.False(t, bytes.Equal(key1, key4))
}

func TestEncryptDecrypt_Success(t *testing.T) {
	plaintext := []byte("Hello, World! This is a secret message.")
	password := "strongpassword123"

	encrypted, err := Encrypt(plaintext, password)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := Decrypt(encrypted, password)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecrypt_EmptyPlaintext(t *testing.T) {
	plaintext := []byte("")
	password := "password"

	encrypted, err := Encrypt(plaintext, password)
	require.NoError(t, err)

	decrypted, err := Decrypt(encrypted, password)
	require.NoError(t, err)
	assert.Empty(t, decrypted) // вместо assert.Equal
}

func TestEncryptDecrypt_LargePlaintext(t *testing.T) {
	plaintext := make([]byte, 1024*1024) // 1 MB
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}
	password := "password"

	encrypted, err := Encrypt(plaintext, password)
	require.NoError(t, err)

	decrypted, err := Decrypt(encrypted, password)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDecrypt_WrongPassword(t *testing.T) {
	plaintext := []byte("secret data")
	password := "correctpassword"

	encrypted, err := Encrypt(plaintext, password)
	require.NoError(t, err)

	_, err = Decrypt(encrypted, "wrongpassword")
	require.Error(t, err)
}

func TestDecrypt_TooShort(t *testing.T) {
	_, err := Decrypt([]byte("short"), "password")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

func TestDecrypt_CorruptedData(t *testing.T) {
	plaintext := []byte("secret data")
	password := "password"

	encrypted, err := Encrypt(plaintext, password)
	require.NoError(t, err)

	// Портим данные
	encrypted[len(encrypted)-1] ^= 0xFF

	_, err = Decrypt(encrypted, password)
	require.Error(t, err)
}

func TestEncrypt_ProducesDifferentCiphertext(t *testing.T) {
	plaintext := []byte("same plaintext")
	password := "password"

	encrypted1, err := Encrypt(plaintext, password)
	require.NoError(t, err)

	encrypted2, err := Encrypt(plaintext, password)
	require.NoError(t, err)

	// Из-за случайных salt и nonce шифротексты должны быть разными
	assert.False(t, bytes.Equal(encrypted1, encrypted2))

	// Но оба должны расшифровываться
	decrypted1, err := Decrypt(encrypted1, password)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted1)

	decrypted2, err := Decrypt(encrypted2, password)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}

func TestEncryptDecrypt_SpecialCharacters(t *testing.T) {
	testCases := []struct {
		name      string
		plaintext []byte
		password  string
	}{
		{"Unicode", []byte("Привет, мир! 🔐"), "пароль123"},
		{"Binary", []byte{0x00, 0x01, 0xFF, 0xFE}, "password"},
		{"Newlines", []byte("line1\nline2\r\nline3"), "password"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := Encrypt(tc.plaintext, tc.password)
			require.NoError(t, err)

			decrypted, err := Decrypt(encrypted, tc.password)
			require.NoError(t, err)
			assert.Equal(t, tc.plaintext, decrypted)
		})
	}
}
