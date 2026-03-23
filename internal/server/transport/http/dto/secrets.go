package dto

// CreateSecretRequest — запрос на создание секрета.
type CreateSecretRequest struct {
	Name     string            `json:"name"`
	Type     int               `json:"type"`
	Data     []byte            `json:"data"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// UpdateSecretRequest — запрос на обновление секрета.
type UpdateSecretRequest struct {
	Name     string            `json:"name"`
	Type     int               `json:"type"`
	Data     []byte            `json:"data"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SecretResponse — ответ с данными секрета.
type SecretResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      int               `json:"type"`
	Data      []byte            `json:"data"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Version   int64             `json:"version"`
	UpdatedAt int64             `json:"updated_at"`
	IsDeleted bool              `json:"is_deleted,omitempty"`
}

// SecretListResponse — ответ со списком секретов.
type SecretListResponse struct {
	Secrets []SecretResponse `json:"secrets"`
}
