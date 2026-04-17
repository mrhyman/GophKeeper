package models

type SecretType int

const (
	SecretTypeLoginPassword SecretType = iota + 1
	SecretTypeTextData
	SecretTypeBinaryData
	SecretTypeBankCard
)

func (s SecretType) String() string {
	switch s {
	case SecretTypeLoginPassword:
		return "login_password"
	case SecretTypeTextData:
		return "text_data"
	case SecretTypeBinaryData:
		return "binary_data"
	case SecretTypeBankCard:
		return "bank_card"
	default:
		return "unknown"
	}
}

type Secret struct {
	ID        string
	UserID    string
	Name      string
	Type      SecretType
	Data      []byte            // зашифрованный blob
	Metadata  map[string]string // произвольная метаинформация
	Version   int64
	UpdatedAt int64
	IsDeleted bool
}
