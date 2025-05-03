package main

import (
	"go_redis/internals/server"
	"go_redis/internals/store"
	"log"
	"time"
)

func main() {
	s := store.NewStore()
	s.StartCleaner(1 * time.Second)

	srv := server.NewServer(":6379", s)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}