package server

import (
	"cicada"
	"cicada/internal/server/store/chat"
	"cicada/internal/server/store/room"
	"context"
	"encoding/json"
	"errors"
	uuid "github.com/satori/go.uuid"
	"log/slog"
	"nhooyr.io/websocket"
	"sync"
	"time"
)

type ChatService struct {
	m       *sync.Mutex
	clients map[string]chan []byte
	cs      *chat.Store
	rs      *room.Store
}

func New(cs *chat.Store, rs *room.Store) *ChatService {
	return &ChatService{
		m:       &sync.Mutex{},
		clients: make(map[string]chan []byte),
		cs:      cs,
		rs:      rs,
	}
}

func (s *ChatService) Disconnect(userId string) {
	delete(s.clients, userId)
}

func (s *ChatService) CreateRoom(r cicada.Room) (cicada.Room, error) {
	id, err := s.rs.Create(r)
	if err != nil {
		return cicada.Room{}, err
	}
	r.Id = id
	return r, nil
}

func (s *ChatService) JoinRoom(userId, roomId string) ([]cicada.ChatMessage, error) {
	s.m.Lock()
	defer s.m.Unlock()

	r, err := s.rs.Get(roomId)
	if err != nil {
		return nil, err
	}

	r.Members = append(r.Members, userId)
	err = s.rs.Update(r)
	if err != nil {
		return nil, err
	}

	go s.SendMessage(systemMessage(roomId, userId+" has joined"))

	return s.cs.GetWindow(roomId, 0, 100)
}

func (s *ChatService) LeaveRoom(userId, roomId string) error {
	s.m.Lock()
	defer s.m.Unlock()

	// check that there is a user id
	r, err := s.rs.Get(roomId)
	if err != nil {
		return err
	}

	found := false
	for i, m := range r.Members {
		if m == userId {
			found = true
			r.Members = append(r.Members[:i], r.Members[i+1:]...) // remove the user from the members list
			break
		}
	}

	if !found {
		return errors.New("member not in room")
	}

	// if there are no more users in the room, delete the room and the chat associated with it
	if len(r.Members) == 0 {
		err = s.rs.Delete(roomId)
	} else {
		delete(s.clients, userId)
		err = s.rs.Update(r)
		go s.SendMessage(systemMessage(roomId, userId+" left the room")) // send notification
	}

	return err
}

func systemMessage(roomId, text string) cicada.ChatMessage {
	return cicada.ChatMessage{
		Id:     uuid.NewV4().String(),
		Date:   time.Now(),
		RoomId: roomId,
		Sender: "system",
		Text:   text,
	}
}

func (s *ChatService) SendMessage(m cicada.ChatMessage) error {
	r, err := s.rs.Get(m.RoomId)
	if err != nil {
		return errors.New("no room with id " + m.RoomId)
	}

	err = s.cs.Save(m)
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	for _, uid := range r.Members {
		ch, ok := s.clients[uid]
		if !ok {
			continue
		}
		ch <- bytes
	}
	return nil
}

func (s *ChatService) Connect(userId string, ws *websocket.Conn) {
	ch := make(chan []byte)
	s.clients[userId] = ch
	go writeLoop(ch, ws) // TODO: this needs a done channel & context
}

func writeLoop(ch chan []byte, ws *websocket.Conn) {
	for {
		select {
		case bytes := <-ch:
			err := ws.Write(context.Background(), websocket.MessageText, bytes)
			if err != nil {
				slog.Error("error writing to client", err)
				break
			}
		}
	}
}
