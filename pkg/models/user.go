package models

type User struct {
	ID           string
	Login        string
	PasswordHash string
	CreatedAt    int64
}
