package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_Struct(t *testing.T) {
	user := User{
		ID:           "user-123",
		Login:        "testuser",
		PasswordHash: "hashed-password",
		CreatedAt:    1234567890,
	}

	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "testuser", user.Login)
	assert.Equal(t, "hashed-password", user.PasswordHash)
	assert.Equal(t, int64(1234567890), user.CreatedAt)
}

func TestUser_ZeroValue(t *testing.T) {
	var user User

	assert.Empty(t, user.ID)
	assert.Empty(t, user.Login)
	assert.Empty(t, user.PasswordHash)
	assert.Zero(t, user.CreatedAt)
}
