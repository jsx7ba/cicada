package main

import (
	"bytes"
	"cicada"
	"cicada/internal/server"
	"cicada/internal/server/store/image"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"nhooyr.io/websocket"
)

type roomRequest struct {
	RoomId string `json:"roomId"`
	UserId string `json:"userId"`
}

type registerMessage struct {
	UserId string `json:"userId"`
}

type HttpHandler struct {
	cs         *server.ChatService
	imageStore *image.Store
}

func (h *HttpHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	imageId := r.PathValue("id")
	b, err := h.imageStore.Get(imageId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	_, err = io.Copy(w, bytes.NewReader(b))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *HttpHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	room := cicada.Room{}
	err := processJsonRequest(r, &room)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newRoom, err := h.cs.CreateRoom(room)
	writeJsonResponse(w, newRoom)
}

func (h *HttpHandler) Room(w http.ResponseWriter, r *http.Request) {
	roomReq := roomRequest{}
	err := processJsonRequest(r, &roomReq)
	if err != nil {
		responseFromError(err, w)
		return
	}

	action := r.URL.Query().Get("a")
	if len(action) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if action == "join" {
		err = h.cs.LeaveRoom(roomReq.UserId, roomReq.RoomId)
	} else if action == "roomReq" {
		chatLog, err := h.cs.JoinRoom(roomReq.UserId, roomReq.RoomId)
		if err == nil {
			writeJsonResponse(w, chatLog)
		}
	} else {
		if len(action) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if err != nil {
		responseFromError(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *HttpHandler) Connect(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{"cicada_v1"},
	})

	if err != nil {
		slog.Error("accept", "error", err)
		return
	}

	mtype, bytes, err := c.Read(context.Background())
	if mtype != websocket.MessageText {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	register := registerMessage{}
	err = json.Unmarshal(bytes, &register)
	if err != nil {
		slog.Error("error decoding registration", "error", err)
		return
	}

	h.cs.Connect(register.UserId, c)
	slog.Info("new connection", "user", register.UserId)
}

func (h *HttpHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	mesg := registerMessage{}
	err := processJsonRequest(r, &mesg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (h *HttpHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	mesg := cicada.ChatMessage{}
	err := processJsonRequest(r, &mesg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
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

func writeJsonResponse(w http.ResponseWriter, value interface{}) {
	bytes, err := json.Marshal(value)
	if err != nil {
		w.WriteHeader(500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, err = w.Write(bytes)
	if err != nil {
		slog.Error("failed to write json response", err)
	}
}

func responseFromError(e error, w http.ResponseWriter) {
	code := http.StatusInternalServerError
	if errors.Is(e, cicada.ErrorNotFound) {
		code = http.StatusNotFound
	} else if errors.Is(e, cicada.ErrorBadRequest) {
		code = http.StatusNotFound
	}

	w.WriteHeader(code)
}
