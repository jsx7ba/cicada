package image

import (
	"cicada"
	"crypto/sha256"
	"encoding/hex"
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
		key := []byte(id)
		_, err := txn.Get(key)
		if err != nil {
			return err
		}
		return txn.Delete(key)
	})
	return processError(err)
}

func (r *Store) Put(bytes []byte) (string, error) {
	hasher := sha256.New()
	hasher.Write(bytes)
	id := hex.EncodeToString(hasher.Sum(nil))
	err := r.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(id), bytes)
	})

	if err != nil {
		return "", err
	}
	return id, nil
}

func processError(e error) error {
	if e == nil {
		return e
	}

	if errors.Is(e, badger.ErrKeyNotFound) {
		return cicada.ErrorNotFound
	}

	return e
}
