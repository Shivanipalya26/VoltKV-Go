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

	if err := server.Start(":6379", s); err != nil {
		log.Fatal(err)
	}
}