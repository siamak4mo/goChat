package main

import (
	"log"
	"server/server"
)

func main() {
	log.Println("INITIALIZING the server")

	s := server.New()
	err := s.Serve()

	if err != nil {
		log.Fatalf("Could not listen -- addr: %s\n", s.Conf.Server.Laddr)
	}
}
