package image

import (
	"bytes"
	"crypto/rand"
	"github.com/dgraph-io/badger/v4"
	"log"
	"log/slog"
	"os"
	"testing"
)

func database() (*badger.DB, string) {
	dbDir := os.TempDir()
	slog.Info("using temp dir", "dir", dbDir)
	db, err := badger.Open(badger.DefaultOptions(dbDir))
	if err != nil {
		log.Fatal("unable to open database", err)
	}
	return db, dbDir
}

func TestRoundTrip(t *testing.T) {
	db, dbDir := database()
	defer os.RemoveAll(dbDir)
	defer db.Close()

	image := make([]byte, 1024)
	_, err := rand.Read(image)
	if err != nil {
		t.Fatal("error generating random bytes", err)
	}

	store := NewStore(db)
	id, err := store.Put(image)
	if err != nil {
		t.Fatal("error writing image to store", err)
	}

	if len(id) == 0 {
		t.Fatal("zero length id returned from Put")
	}

	newImage, err := store.Get(id)
	if err != nil {
		t.Fatal("error writing image to store", err)
	}

	if len(newImage) != len(image) {
		t.Error("failed to read as many bytes as was written")
	}
	if bytes.Compare(image, newImage) != 0 {
		t.Error("image and new image are not the same")
	}
}

func TestDelete(t *testing.T) {
	db, dbDir := database()
	defer os.RemoveAll(dbDir)
	defer db.Close()

	image := make([]byte, 1024)
	_, err := rand.Read(image)
	if err != nil {
		t.Fatal("error generating random bytes", err)
	}

	store := NewStore(db)
	id, err := store.Put(image)
	if err != nil {
		t.Fatal("error writing image to store", err)
	}

	if len(id) == 0 {
		t.Fatal("zero length id returned from Put")
	}

	err = store.Delete(id)
	if err != nil {
		t.Error("failed to delete image", err)
	}
}
