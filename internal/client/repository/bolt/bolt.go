package bolt

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketSecrets = []byte("secrets")
	bucketAuth    = []byte("auth")
)

type DB struct {
	db *bolt.DB
}

// New открывает (или создаёт) файл bbolt и инициализирует бакеты.
func New(path string) (*DB, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1})
	if err != nil {
		return nil, fmt.Errorf("open bolt: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucketSecrets, bucketAuth} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return fmt.Errorf("create bucket %s: %w", b, err)
			}
		}
		return nil
	})
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &DB{db: db}, nil
}

// Close закрывает файл bbolt.
func (d *DB) Close() error {
	return d.db.Close()
}
