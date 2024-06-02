package chat

import (
	"cicada"
	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
	"log"
)

const (
	collection = "chats"
)

type Store struct {
	db *clover.DB
}

func NewStore(db *clover.DB) *Store {
	exists, err := db.HasCollection(collection)
	if err != nil {
		log.Fatal("failed to create collection", collection, err)
	}

	if !exists {
		err := db.CreateCollection(collection)
		if err != nil {
			log.Fatal("failed to create collection", collection, err)
		}
		err = db.CreateIndex(collection, "roomId")
		if err != nil {
			log.Fatal("failed to create roomId index for collection:", collection, err)
		}
		err = db.CreateIndex(collection, "date")
		if err != nil {
			log.Fatal("failed to create date index for collection:", collection, err)
		}
	}
	return &Store{db: db}
}

func (s *Store) Save(m cicada.ChatMessage) error {
	doc := document.NewDocumentOf(m)
	return s.db.Insert(collection, doc)
}

// GetWindow fetches a page of chat messages, sorted by date.
func (s *Store) GetWindow(roomId string, from, size int) ([]cicada.ChatMessage, error) {
	if from < 0 || size <= 0 {
		return nil, cicada.ErrorBadRequest
	}

	q := query.NewQuery(collection).
		Skip(from).
		Limit(size).
		Sort(query.SortOption{Field: "date", Direction: 1}).
		Where(query.Field("roomId").Eq(roomId))

	docs, err := s.db.FindAll(q)
	if err != nil {
		return nil, err
	}

	messages := make([]cicada.ChatMessage, len(docs))
	for i, m := range docs {
		err := m.Unmarshal(&messages[i])
		if err != nil {
			return nil, err
		}
	}
	return messages, nil
}
