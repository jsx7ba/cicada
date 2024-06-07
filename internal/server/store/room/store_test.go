package room

import (
	"cicada"
	"errors"
	"github.com/ostafen/clover/v2"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"
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

func TestRoundTrip(t *testing.T) {
	db, dir := database()
	defer db.Close()
	defer os.Remove(filepath.Join(dir, "data.db"))

	store := NewStore(db)

	room := cicada.Room{
		Name:        "Water Cooler",
		Description: "Idle chit chat",
		Members:     []string{"user1", "user2"},
	}

	id, err := store.Put(room)
	if err != nil {
		t.Fatal("error saving room", err)
	}

	if len(id) == 0 {
		t.Fatal("generated id has no length")
	}

	newRoom, err := store.Get(id)
	if err != nil {
		t.Fatal("error fetching room "+id, err)
	}

	if newRoom.Id != id {
		t.Errorf("id was not correctly saved, expected '%s' got '%s'", id, newRoom.Id)
	}

	if newRoom.Name != room.Name {
		t.Errorf("name didn't round trip, expected '%s' got '%s'", room.Name, newRoom.Name)
	}

	if newRoom.Description != room.Description {
		t.Errorf("description didn't round trip, expected '%s' got '%s'", room.Name, newRoom.Name)
	}

	if len(newRoom.Members) != len(room.Members) {
		t.Errorf("members didn't round trip, expected '%+v' got '%+v'", room.Members, newRoom.Members)
	}

	if !reflect.DeepEqual(room.Members, newRoom.Members) {
		t.Errorf("members didn't round trip, expected '%+v' got '%+v'", room.Members, newRoom.Members)
	}
}

func TestUpdate(t *testing.T) {
	db, dir := database()
	defer db.Close()
	defer os.Remove(filepath.Join(dir, "data.db"))

	store := NewStore(db)

	room := cicada.Room{
		Name:        "Water Cooler",
		Description: "Idle chit chat",
		Members:     []string{"user1", "user2"},
	}

	id, err := store.Put(room)
	if err != nil {
		t.Fatal("error saving room", err)
	}

	update := cicada.Room{
		Id:          id,
		Name:        room.Name,
		Description: "Idle Talk",
		Members:     room.Members,
	}

	err = store.Update(update)
	if err != nil {
		t.Fatal("failed to update room", err)
	}

}

func TestForUser(t *testing.T) {
	db, dir := database()
	defer db.Close()
	defer os.Remove(filepath.Join(dir, "data.db"))

	store := NewStore(db)

	rooms := []cicada.Room{
		{
			Name:        "Water Cooler",
			Description: "Idle chit chat",
			Members:     []string{"user1", "user2"},
		},
		{
			Name:        "MOTD",
			Description: "Fun message of the day",
			Members:     []string{"user1", "user2", "user3"},
		},
		{
			Name:        "Service Design",
			Description: "Design discussion about the new service",
			Members:     []string{"user13", "user2", "user21", "user34"},
		},
	}

	for _, r := range rooms {
		_, err := store.Put(r)
		if err != nil {
			t.Fatal("error saving room", err)
		}
	}

	testForUser(t, store, "user1", 2)
	testForUser(t, store, "user2", 3)
	testForUser(t, store, "user3", 1)
	testForUser(t, store, "user34", 1)
}

func testForUser(t *testing.T, store *Store, uid string, expected int) {
	forUser, err := store.GetForUser(uid)
	if err != nil {
		t.Fatal("retrieving rooms for userid1 failed", err)
	}

	if len(forUser) != expected {
		t.Errorf("GetForUser returned %d rooms, expected %d", len(forUser), expected)
	}
}

func TestGetBad(t *testing.T) {
	db, dir := database()
	defer db.Close()
	defer os.Remove(filepath.Join(dir, "data.db"))

	store := NewStore(db)
	_, err := store.Get("asdfasdf")
	if err == nil {
		t.Error("expected error got none")
	} else if !errors.Is(err, cicada.ErrorNotFound) {
		t.Error("expected error not found, got ", err)
	}
}

func TestGetAll(t *testing.T) {
	db, dir := database()
	defer db.Close()
	defer os.Remove(filepath.Join(dir, "data.db"))

	store := NewStore(db)

	rooms1 := []cicada.Room{
		{
			Name:        "Water Cooler",
			Description: "Idle chit chat",
			Members:     []string{"user1", "user2"},
		},
		{
			Name:        "MOTD",
			Description: "Fun message of the day",
			Members:     []string{"user1", "user2", "user3"},
		},
		{
			Name:        "Service Design",
			Description: "Design discussion about the new service",
			Members:     []string{"user13", "user2", "user21", "user34"},
		},
	}

	for i, v := range rooms1 {
		id, err := store.Put(v)
		if err != nil {
			t.Fatal("unable to store room")
		}
		rooms1[i].Id = id
	}

	rooms2, err := store.GetAll()
	if err != nil {
		t.Fatal("error fetching all rooms", err)
	}

	if len(rooms2) != len(rooms1) {
		t.Errorf("Expected %d rooms got %d", len(rooms1), len(rooms2))
	}

	oldRooms := mapRooms(rooms1)
	newRooms := mapRooms(rooms2)

	for k, r1 := range oldRooms {
		r2 := newRooms[k]
		if r1.Name != r2.Name {
			t.Errorf("name mismatch, expected %s, got %s", r1.Name, r2.Name)
		}
		if r1.Description != r2.Description {
			t.Errorf("description mismatch, expected %s, got %s", r1.Description, r2.Description)
		}
		if !reflect.DeepEqual(r1.Members, r2.Members) {
			t.Errorf("members mismatch, expected %s, got %s", r1.Description, r2.Description)
		}
	}
}

func mapRooms(rooms []cicada.Room) map[string]cicada.Room {
	m := make(map[string]cicada.Room)
	for _, r := range rooms {
		m[r.Id] = r
	}
	return m
}

func TestGetAllEmpty(t *testing.T) {
	db, dir := database()
	defer db.Close()
	defer os.Remove(filepath.Join(dir, "data.db"))

	store := NewStore(db)
	rooms, err := store.GetAll()
	if err != nil {
		t.Fatal("get all errored", err)
	}

	if len(rooms) != 0 {
		t.Errorf("somehow a db with no rooms return %d rooms", len(rooms))
	}
}

// clover doesn't seem to return any way to surmise if delete was successful.
func TestDeleteBad(t *testing.T) {
	db, dir := database()
	defer db.Close()
	defer os.Remove(filepath.Join(dir, "data.db"))

	store := NewStore(db)
	err := store.Delete("asdfasdf")
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	if !errors.Is(err, cicada.ErrorNotFound) {
		t.Error("expected not found, got ", err)
	}

}
