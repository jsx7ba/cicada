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
	"strconv"
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
	defer db.Close()
	defer os.Remove(filepath.Join(dir, "data.db"))

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

func TestPagingEmpty(t *testing.T) {
	db, dir := database()
	defer db.Close()
	defer os.Remove(filepath.Join(dir, "data.db"))

	s := NewStore(db)
	messages := generateMessages("237", 10)

	for i := 0; i != 10; i++ {
		err := s.Save(messages[i])
		if err != nil {
			t.Fatal("unable to save messages", err)
		}
	}
	for i := 0; i != 100; i++ {
		cm, err := s.GetWindow("237", i, 1)
		if err != nil {
			t.Fatal("unable get messages", err)
		}
		if i < 10 && len(cm) != 1 {
			t.Fatal("slice length didn't match the window size")
		} else if i > 10 && len(cm) != 0 {
			t.Fatal("slice length should be zero")
		}
	}
}

func TestPaging(t *testing.T) {
	db, dir := database()
	defer db.Close()
	defer os.Remove(filepath.Join(dir, "data.db"))

	s := NewStore(db)
	messages := generateMessages("237", 10)

	for i := 0; i != 10; i++ {
		err := s.Save(messages[i])
		if err != nil {
			t.Fatal("unable to save messages", err)
		}
	}

	for i := 0; i != 10; i++ {
		cm, err := s.GetWindow("237", i, 1)
		if err != nil {
			t.Fatal("unable get messages", err)
		}
		if len(cm) != 1 {
			t.Fatal("slice length didn't match the window size")
		}
		fmt.Println(cm[0].Text)
	}

}

func generateMessages(roomId string, count int) []cicada.ChatMessage {
	now := time.Now()
	messages := make([]cicada.ChatMessage, count)
	for i := 0; i != count; i++ {
		m := cicada.ChatMessage{
			Id:     uuid.NewV4().String(),
			Date:   now.Add(time.Second + time.Duration(i)),
			RoomId: roomId,
			Sender: "justin@justin.com",
			Text:   "the crow flies at midnight" + strconv.Itoa(i),
		}
		messages[i] = m
	}
	return messages
}
