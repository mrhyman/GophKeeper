package screens

import (
	"gophkeeper/pkg/models"
)

// NavigateMsg — команда перехода на другой экран.
type NavigateMsg struct {
	Screen     int
	Secret     *models.Secret
	SecretType models.SecretType
	Filter     models.SecretType
	MasterPass string // передаётся при логине для инициализации шифрования
}

type LogoutReason int

const (
	LogoutReasonUser LogoutReason = iota
	LogoutReasonTokenExpired
	LogoutReasonServerError
)

// ErrorMsg — сообщение об ошибке.
type ErrorMsg struct {
	Err error
}

// StatusMsg — сообщение статуса.
type StatusMsg struct {
	Text string
}

type LogoutMsg struct {
	Reason  LogoutReason
	Message string
}

type DeleteSecretMsg struct {
	SecretID string
}
