package main

import (
	"log"
	"net/http"
	"os"

	"github.com/edgar981/jukebox-queue/go/api"
	"github.com/edgar981/jukebox-queue/go/queue"
)

func main() {
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("jukebox-queue (Go) listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, api.NewHandler(queue.NewStore(nil))))
}
