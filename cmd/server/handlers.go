package main

import (
	"bytes"
	"cicada"
	"cicada/internal/server"
	"cicada/internal/server/store/image"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"nhooyr.io/websocket"
)

type HttpHandler struct {
	cs         *server.ChatService
	imageStore *image.Store
}

func (m *HttpHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	imageId := r.PathValue("id")
	b, err := m.imageStore.Get(imageId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	_, err = io.Copy(w, bytes.NewReader(b))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func (m *HttpHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {

}

func (m *HttpHandler) Unregister(w http.ResponseWriter, r *http.Request) {
	mesg := cicada.RegisterMessage{}
	err := processJsonRequest(r, &mesg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (m *HttpHandler) ProcessMessage(w http.ResponseWriter, r *http.Request) {
	mesg := cicada.ChatMessage{}
	err := processJsonRequest(r, &mesg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func (m *HttpHandler) Register(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{"cicada_v1"},
	})

	if err != nil {
		slog.Error("accept", "error", err)
		return
	}

	mtype, bytes, err := c.Read(context.Background())
	if mtype != websocket.MessageText {
		//return errors.NewStore("expected text response from websocket")
	}

	register := cicada.RegisterMessage{}
	err = json.Unmarshal(bytes, &register)
	if err != nil {
		//return fmt.Errorf("error decoding registration %w", err)
	}

	m.cs.Connect(register.UserId, c)
}

func processJsonRequest[E any](r *http.Request, value *E) error {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	err = json.Unmarshal(bytes, value)
	return err
}
