package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretType_String(t *testing.T) {
	tests := []struct {
		name       string
		secretType SecretType
		want       string
	}{
		{
			name:       "login password",
			secretType: SecretTypeLoginPassword,
			want:       "login_password",
		},
		{
			name:       "text data",
			secretType: SecretTypeTextData,
			want:       "text_data",
		},
		{
			name:       "binary data",
			secretType: SecretTypeBinaryData,
			want:       "binary_data",
		},
		{
			name:       "bank card",
			secretType: SecretTypeBankCard,
			want:       "bank_card",
		},
		{
			name:       "unknown type",
			secretType: SecretType(999),
			want:       "unknown",
		},
		{
			name:       "zero value",
			secretType: SecretType(0),
			want:       "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.secretType.String()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSecretType_Constants(t *testing.T) {
	assert.Equal(t, SecretType(1), SecretTypeLoginPassword)
	assert.Equal(t, SecretType(2), SecretTypeTextData)
	assert.Equal(t, SecretType(3), SecretTypeBinaryData)
	assert.Equal(t, SecretType(4), SecretTypeBankCard)
}

func TestSecret_Struct(t *testing.T) {
	secret := Secret{
		ID:        "test-id",
		UserID:    "user-123",
		Name:      "my-secret",
		Type:      SecretTypeLoginPassword,
		Data:      []byte("encrypted-data"),
		Metadata:  map[string]string{"key": "value"},
		Version:   1,
		UpdatedAt: 1234567890,
		IsDeleted: false,
	}

	assert.Equal(t, "test-id", secret.ID)
	assert.Equal(t, "user-123", secret.UserID)
	assert.Equal(t, "my-secret", secret.Name)
	assert.Equal(t, SecretTypeLoginPassword, secret.Type)
	assert.Equal(t, []byte("encrypted-data"), secret.Data)
	assert.Equal(t, "value", secret.Metadata["key"])
	assert.Equal(t, int64(1), secret.Version)
	assert.Equal(t, int64(1234567890), secret.UpdatedAt)
	assert.False(t, secret.IsDeleted)
}
