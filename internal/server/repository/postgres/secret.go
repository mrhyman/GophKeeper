package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"gophkeeper/pkg/models"
)

var ErrSecretNotFound = errors.New("secret not found")

// SecretRepository реализует repository.SecretRepository для PostgreSQL.
type SecretRepository struct {
	db *DB
}

// NewSecretRepository создаёт новый SecretRepository.
func NewSecretRepository(db *DB) *SecretRepository {
	return &SecretRepository{db: db}
}

// Create создаёт новый секрет.
func (r *SecretRepository) Create(ctx context.Context, secret *models.Secret) error {
	meta, err := json.Marshal(secret.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO secrets (user_id, name, type, data, metadata, version, updated_at)
		VALUES ($1, $2, $3, $4, $5, 1, $6)
		RETURNING id
	`

	now := time.Now().Unix()

	err = r.db.Pool.QueryRow(ctx, query,
		secret.UserID,
		secret.Name,
		int(secret.Type),
		secret.Data,
		meta,
		now,
	).Scan(&secret.ID)
	if err != nil {
		return fmt.Errorf("create secret: %w", err)
	}

	secret.Version = 1
	secret.UpdatedAt = now

	return nil
}

// GetByID возвращает секрет по ID с проверкой владельца.
func (r *SecretRepository) GetByID(ctx context.Context, userID, secretID string) (*models.Secret, error) {
	query := `
		SELECT id, user_id, name, type, data, metadata, version, updated_at, is_deleted
		FROM secrets
		WHERE id = $1 AND user_id = $2
	`

	secret, err := r.scanSecret(r.db.Pool.QueryRow(ctx, query, secretID, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSecretNotFound
		}
		return nil, fmt.Errorf("get secret by id: %w", err)
	}

	return secret, nil
}

// List возвращает все неудалённые секреты пользователя.
func (r *SecretRepository) List(ctx context.Context, userID string) ([]*models.Secret, error) {
	query := `
		SELECT id, user_id, name, type, data, metadata, version, updated_at, is_deleted
		FROM secrets
		WHERE user_id = $1 AND is_deleted = FALSE
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	defer rows.Close()

	return r.scanSecrets(rows)
}

// Update обновляет секрет с инкрементом версии.
func (r *SecretRepository) Update(ctx context.Context, secret *models.Secret) error {
	meta, err := json.Marshal(secret.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	now := time.Now().Unix()

	query := `
		UPDATE secrets
		SET name       = $1,
		    type       = $2,
		    data       = $3,
		    metadata   = $4,
		    version    = version + 1,
		    updated_at = $5
		WHERE id = $6 AND user_id = $7 AND is_deleted = FALSE
		RETURNING version, updated_at
	`

	err = r.db.Pool.QueryRow(ctx, query,
		secret.Name,
		int(secret.Type),
		secret.Data,
		meta,
		now,
		secret.ID,
		secret.UserID,
	).Scan(&secret.Version, &secret.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSecretNotFound
		}
		return fmt.Errorf("update secret: %w", err)
	}

	return nil
}

// Delete выполняет мягкое удаление секрета.
func (r *SecretRepository) Delete(ctx context.Context, userID, secretID string) error {
	now := time.Now().Unix()

	query := `
		UPDATE secrets
		SET is_deleted  = TRUE,
		    version     = version + 1,
		    updated_at  = $1
		WHERE id = $2 AND user_id = $3 AND is_deleted = FALSE
	`

	ct, err := r.db.Pool.Exec(ctx, query, now, secretID, userID)
	if err != nil {
		return fmt.Errorf("delete secret: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return ErrSecretNotFound
	}

	return nil
}

// GetModifiedAfter возвращает секреты, изменённые после указанного timestamp, включая удаленные.
func (r *SecretRepository) GetModifiedAfter(ctx context.Context, userID string, since int64) ([]*models.Secret, error) {
	query := `
		SELECT id, user_id, name, type, data, metadata, version, updated_at, is_deleted
		FROM secrets
		WHERE user_id = $1 AND updated_at > $2
		ORDER BY updated_at ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, since)
	if err != nil {
		return nil, fmt.Errorf("get modified after: %w", err)
	}
	defer rows.Close()

	return r.scanSecrets(rows)
}

// scanSecret сканирует одну строку в модель Secret.
func (r *SecretRepository) scanSecret(row pgx.Row) (*models.Secret, error) {
	var (
		secret     models.Secret
		secretType int
		metaJSON   []byte
	)

	err := row.Scan(
		&secret.ID,
		&secret.UserID,
		&secret.Name,
		&secretType,
		&secret.Data,
		&metaJSON,
		&secret.Version,
		&secret.UpdatedAt,
		&secret.IsDeleted,
	)
	if err != nil {
		return nil, err
	}

	secret.Type = models.SecretType(secretType)

	if len(metaJSON) > 0 {
		if err := json.Unmarshal(metaJSON, &secret.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	return &secret, nil
}

// scanSecrets сканирует набор строк в слайс Secret.
func (r *SecretRepository) scanSecrets(rows pgx.Rows) ([]*models.Secret, error) {
	var secrets []*models.Secret

	for rows.Next() {
		var (
			secret     models.Secret
			secretType int
			metaJSON   []byte
		)

		err := rows.Scan(
			&secret.ID,
			&secret.UserID,
			&secret.Name,
			&secretType,
			&secret.Data,
			&metaJSON,
			&secret.Version,
			&secret.UpdatedAt,
			&secret.IsDeleted,
		)
		if err != nil {
			return nil, fmt.Errorf("scan secret: %w", err)
		}

		secret.Type = models.SecretType(secretType)

		if len(metaJSON) > 0 {
			if err := json.Unmarshal(metaJSON, &secret.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}

		secrets = append(secrets, &secret)
	}

	return secrets, rows.Err()
}
