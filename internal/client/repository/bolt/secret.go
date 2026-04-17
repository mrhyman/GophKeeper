package bolt

import (
	"encoding/json"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"

	"gophkeeper/pkg/models"
)

var ErrSecretNotFound = errors.New("secret not found")

// SecretRepository реализует repository.SecretRepository для bbolt.
type SecretRepository struct {
	db *DB
}

// NewSecretRepository создаёт новый SecretRepository.
func NewSecretRepository(db *DB) *SecretRepository {
	return &SecretRepository{db: db}
}

// Create сохраняет новый секрет.
func (r *SecretRepository) Create(secret *models.Secret) error {
	return r.db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSecrets)
		return r.put(b, secret)
	})
}

// GetByID возвращает секрет по ID.
func (r *SecretRepository) GetByID(secretID string) (*models.Secret, error) {
	var secret *models.Secret

	err := r.db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSecrets)
		v := b.Get([]byte(secretID))
		if v == nil {
			return ErrSecretNotFound
		}

		var s models.Secret
		if err := json.Unmarshal(v, &s); err != nil {
			return fmt.Errorf("unmarshal secret: %w", err)
		}
		secret = &s
		return nil
	})

	return secret, err
}

// List возвращает все неудалённые секреты.
func (r *SecretRepository) List() ([]*models.Secret, error) {
	var secrets []*models.Secret

	err := r.db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSecrets)

		return b.ForEach(func(k, v []byte) error {
			var s models.Secret
			if err := json.Unmarshal(v, &s); err != nil {
				return fmt.Errorf("unmarshal secret %s: %w", k, err)
			}
			if !s.IsDeleted {
				secrets = append(secrets, &s)
			}
			return nil
		})
	})

	return secrets, err
}

// Update обновляет существующий секрет.
func (r *SecretRepository) Update(secret *models.Secret) error {
	return r.db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSecrets)

		existing := b.Get([]byte(secret.ID))
		if existing == nil {
			return ErrSecretNotFound
		}

		return r.put(b, secret)
	})
}

// Delete помечает секрет как удалённый (soft delete).
func (r *SecretRepository) Delete(secretID string) error {
	return r.db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSecrets)

		v := b.Get([]byte(secretID))
		if v == nil {
			return ErrSecretNotFound
		}

		var s models.Secret
		if err := json.Unmarshal(v, &s); err != nil {
			return fmt.Errorf("unmarshal secret: %w", err)
		}

		s.IsDeleted = true
		s.Version++

		return r.put(b, &s)
	})
}

// UpsertBatch массово создаёт или обновляет секреты.
// Используется при синхронизации — серверная версия перезаписывает локальную.
func (r *SecretRepository) UpsertBatch(secrets []*models.Secret) error {
	return r.db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSecrets)

		for _, secret := range secrets {
			if err := r.put(b, secret); err != nil {
				return fmt.Errorf("upsert secret %s: %w", secret.ID, err)
			}
		}

		return nil
	})
}

// put сериализует и сохраняет секрет в бакет.
func (r *SecretRepository) put(b *bolt.Bucket, secret *models.Secret) error {
	data, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("marshal secret: %w", err)
	}
	return b.Put([]byte(secret.ID), data)
}
