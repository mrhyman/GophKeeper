package dto

// SyncRequest — запрос на синхронизацию.
type SyncRequest struct {
	LastSync      int64            `json:"last_sync"`
	ClientSecrets []SecretResponse `json:"client_secrets,omitempty"`
}

// SyncResponse — ответ синхронизации.
type SyncResponse struct {
	ServerSecrets []SecretResponse `json:"server_secrets"`
	SyncTimestamp int64            `json:"sync_timestamp"`
}
