package cicada

import (
	"time"
)

// ChatMessage A message that the user sends to a room.
type ChatMessage struct {
	Id     string    `clover:"id" json:"id"`
	Date   time.Time `clover:"date" json:"date"`
	RoomId string    `clover:"roomId" json:"roomId"`
	Sender string    `clover:"sender" json:"sender"`
	Text   string    `clover:"text" json:"text"`
	Images []Image   `clover:"text" json:"images"`
}
