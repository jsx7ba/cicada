package main

import (
	"cicada/internal/server"
	"cicada/internal/server/store/chat"
	"cicada/internal/server/store/image"
	"cicada/internal/server/store/room"
	"context"
	badger "github.com/dgraph-io/badger/v4"
	clover "github.com/ostafen/clover/v2"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

func main() {
	err := run()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
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

	quitChan := make(chan interface{})
	chatService := server.New(quitChan, chat.NewStore(objDb), room.NewStore(objDb))

	h := &HttpHandler{
		chatService,
		image.NewStore(kvDb),
	}

	http.HandleFunc("POST /room", h.CreateRoom)
	http.HandleFunc("PUT /room/:id:", h.Room)
	http.HandleFunc("POST /message", h.SendMessage)
	http.HandleFunc("POST /register", h.Connect)
	http.HandleFunc("POST /unregister", h.Disconnect)
	http.HandleFunc("GET /image/:id:", h.GetImage)
	s := &http.Server{}

	errChan := make(chan error)
	go func() {
		errChan <- s.Serve(l)
	}()

	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt)

	select {
	case err := <-errChan:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigs:
		quitChan <- true
		log.Printf("terminating: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return s.Shutdown(ctx)
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
