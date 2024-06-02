package chat

import (
	"cicada"
	"fmt"
	"github.com/ostafen/clover/v2"
	uuid "github.com/satori/go.uuid"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func database() (*clover.DB, string) {
	dbDir := os.TempDir()
	slog.Info("using temp dir", "dir", dbDir)
	db, err := clover.Open(dbDir)
	if err != nil {
		log.Fatal("unable to open database", err)
	}
	return db, dbDir
}

func TestSave(t *testing.T) {
	db, dir := database()
	defer os.Remove(filepath.Join(dir, "data.db"))
	defer db.Close()

	m := cicada.ChatMessage{
		Id:     uuid.NewV4().String(),
		Date:   time.Now(),
		RoomId: "237",
		Sender: "justin@justin.com",
		Text:   "the crow flies at midnight",
	}
	s := NewStore(db)
	err := s.Save(m)
	if err != nil {
		t.Fatal("error occurred while saving a chat message", err)
	}

	w, err := s.GetWindow("237", 0, 1)
	if err != nil {
		t.Fatal("error occurred while getting a chat window", err)
	}

	if len(w) != 1 {
		t.Fatal("expected one chat message, got", len(w))
	}

	m2 := w[0]
	if !m2.Date.Equal(m.Date) {
		t.Error("date did not round trip")
	}
	if m2.RoomId != m.RoomId {
		t.Error("room id did not round trip")
	}
	if m2.Id != m.Id {
		t.Error("id did not round trip")
	}
	if m2.Sender != m.Sender {
		t.Error("sender did not round trip")
	}
	if m2.Text != m.Text {
		t.Error("text did not round trip")
	}

	if t.Failed() {
		fmt.Println("expected: ", m)
		fmt.Println("got:      ", m2)
	}
}
