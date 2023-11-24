package main

import (
	"log"
	"server/server"
	"server/server/config"
)

func main() {
	log.Println("INITIALIZING the server")

	con := config.New()
	s := server.New()
	s.Conf = con

	err := s.Serve()

	if err != nil {
		log.Fatalf("Could not listen -- addr: %s\n", s.Conf.Server.Laddr)
	}
}
