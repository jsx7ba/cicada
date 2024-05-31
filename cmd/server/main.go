package main

import (
	"cicada/internal/server"
	"cicada/internal/server/store/chat"
	"cicada/internal/server/store/image"
	"cicada/internal/server/store/room"
	badger "github.com/dgraph-io/badger/v4"
	clover "github.com/ostafen/clover/v2"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("expected exactly one argument for listen address")
	}

	l, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		log.Fatal("unable to listen on ", os.Args[1])
	}

	kvDb := kvStore("/tmp")
	defer kvDb.Close()

	objDb := objStore("/tmp")
	defer objDb.Close()

	h := &HttpHandler{
		server.New(chat.NewStore(objDb), room.NewStore(objDb)),
		image.NewStore(kvDb),
	}

	//http.HandlerFunc("GET /chatlog", h.ChatLog)
	http.HandleFunc("POST /room", h.CreateRoom)
	http.HandleFunc("POST /message", h.ProcessMessage)
	http.HandleFunc("POST /register", h.Register)
	http.HandleFunc("POST /unregister", h.Unregister)
	http.HandleFunc("GET /image/:id:", h.GetImage)
	s := &http.Server{}
	s.Serve(l)
}

func makeTempDir(prefix, dbName string) string {
	cicadaDir := filepath.Join(prefix, "cicada", dbName)
	err := os.MkdirAll(cicadaDir, 0700)
	checkError("failed to create database directory: "+cicadaDir, err)
	return cicadaDir
}

func objStore(tmpDir string) *clover.DB {
	dbDir := makeTempDir(tmpDir, "clover")
	db, err := clover.Open(dbDir)
	checkError("failed to open clover database", err)
	return db
}

func kvStore(tmpDir string) *badger.DB {
	dbDir := makeTempDir(tmpDir, "badger")
	kv, err := badger.Open(badger.DefaultOptions(dbDir))
	checkError("failed to open badger database", err)
	return kv
}

func checkError(mesg string, err error) {
	if err != nil {
		log.Fatal(mesg, err)
	}
}
