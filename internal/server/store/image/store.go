package image

import (
	"cicada/internal/server/store"
	"crypto/sha256"
	"errors"
	badger "github.com/dgraph-io/badger/v4"
)

type Store struct {
	db *badger.DB
}

func NewStore(db *badger.DB) *Store {
	return &Store{db: db}
}

func (r *Store) Get(id string) ([]byte, error) {
	var buffer []byte

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}

		buffer, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		return nil, processError(err)
	}

	return buffer, nil
}

func (r *Store) Delete(id string) error {
	err := r.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(id))
		return err
	})
	return processError(err)
}

func (r *Store) Create(bytes []byte) (string, error) {
	hasher := sha256.New()
	hasher.Write(bytes)
	id := hasher.Sum(nil)
	err := r.db.Update(func(txn *badger.Txn) error {
		return txn.Set(id, bytes)
	})

	if err != nil {
		return "", err
	}
	return string(id), nil
}

func processError(e error) error {
	if e == nil {
		return e
	}

	if errors.Is(e, badger.ErrKeyNotFound) {
		return store.ErrorNotFound
	}

	return e
}
