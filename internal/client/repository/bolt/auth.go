package bolt

import (
	"encoding/binary"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

var (
	keyAccessToken  = []byte("access_token")
	keyRefreshToken = []byte("refresh_token")
	keyLastSync     = []byte("last_sync")
)

var ErrNoTokens = errors.New("no tokens stored")

// AuthRepository реализует repository.AuthRepository для bbolt.
type AuthRepository struct {
	db *DB
}

// NewAuthRepository создаёт новый AuthRepository.
func NewAuthRepository(db *DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// SaveTokens сохраняет пару токенов.
func (r *AuthRepository) SaveTokens(access, refresh string) error {
	return r.db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAuth)

		if err := b.Put(keyAccessToken, []byte(access)); err != nil {
			return fmt.Errorf("put access token: %w", err)
		}
		if err := b.Put(keyRefreshToken, []byte(refresh)); err != nil {
			return fmt.Errorf("put refresh token: %w", err)
		}

		return nil
	})
}

// GetTokens возвращает сохранённые токены.
func (r *AuthRepository) GetTokens() (string, string, error) {
	var access, refresh string

	err := r.db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAuth)

		a := b.Get(keyAccessToken)
		rf := b.Get(keyRefreshToken)

		if len(a) == 0 || len(rf) == 0 {
			return ErrNoTokens
		}

		access = string(a)
		refresh = string(rf)

		return nil
	})

	return access, refresh, err
}

// DeleteTokens удаляет токены (логаут).
func (r *AuthRepository) DeleteTokens() error {
	return r.db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAuth)
		_ = b.Delete(keyAccessToken)
		_ = b.Delete(keyRefreshToken)
		return nil
	})
}

// SaveLastSync сохраняет timestamp последней синхронизации.
func (r *AuthRepository) SaveLastSync(timestamp int64) error {
	return r.db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAuth)
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(timestamp))
		return b.Put(keyLastSync, buf)
	})
}

// GetLastSync возвращает timestamp последней синхронизации.
func (r *AuthRepository) GetLastSync() (int64, error) {
	var ts int64

	err := r.db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketAuth)
		v := b.Get(keyLastSync)
		if v == nil {
			return nil // 0 — синхронизации не было
		}
		ts = int64(binary.BigEndian.Uint64(v))
		return nil
	})

	return ts, err
}
