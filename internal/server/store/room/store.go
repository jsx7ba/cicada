package room

import (
	"cicada"
	"errors"
	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
	uuid "github.com/satori/go.uuid"
	"log"
)

const (
	collection = "rooms"
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

		err = db.CreateIndex(collection, "members")
		if err != nil {
			log.Fatal("failed to create members index for collection:", collection, err)
		}

		err = db.CreateIndex(collection, "id")
		if err != nil {
			log.Fatal("failed to create id index for collection:", collection, err)
		}
	}
	return &Store{db: db}
}

func (s *Store) Put(r cicada.Room) (string, error) {
	docId := uuid.NewV4().String()
	r.Id = docId
	doc := document.NewDocumentOf(r)
	_, err := s.db.InsertOne(collection, doc)

	if err != nil {
		docId = ""
	}

	return docId, err
}

func (s *Store) Update(r cicada.Room) error {
	q := query.NewQuery(collection).Where(query.Field("id").Eq(r.Id))
	doc := document.NewDocumentOf(r)
	err := s.db.Update(q, doc.AsMap())
	return processError(err)
}

func (s *Store) GetForUser(id string) ([]cicada.Room, error) {
	q := query.NewQuery(collection).Where(query.Field("members").Contains(id))
	docs, err := s.db.FindAll(q)
	if e := processError(err); e != nil {
		return nil, e
	}
	rooms := make([]cicada.Room, len(docs))
	for i, d := range docs {
		if err = d.Unmarshal(&rooms[i]); err != nil {
			return nil, err
		}
	}
	return rooms, nil
}

func (s *Store) Delete(id string) error {
	q := query.NewQuery(collection).Where(query.Field("id").Eq(id))
	return processError(s.db.Delete(q))
}

func (s *Store) Get(id string) (cicada.Room, error) {
	r := cicada.Room{}
	q := query.NewQuery(collection).Where(query.Field("id").Eq(id))
	doc, err := s.db.FindFirst(q)
	if doc == nil {
		return r, cicada.ErrorNotFound
	}

	if e := processError(err); e != nil {
		return r, e
	}
	err = doc.Unmarshal(&r)
	return r, err
}

func (s *Store) GetAll() ([]cicada.Room, error) {
	q := query.NewQuery(collection)
	docs, err := s.db.FindAll(q)
	if e := processError(err); e != nil {
		return nil, e
	}

	rooms := make([]cicada.Room, len(docs))
	for i, d := range docs {
		r := cicada.Room{}
		err = d.Unmarshal(&r)
		if err != nil {
			break
		}
		rooms[i] = r
	}
	return rooms, err
}

func processError(e error) error {
	if e == nil {
		return nil
	}

	if errors.Is(e, clover.ErrDocumentNotExist) {
		return cicada.ErrorNotFound
	}

	return e
}
