package main

import (
	"log"
	"server/server"
	"server/server/config"
	"sync"
)

var (
	chat_s *server.Server
	conf   *config.Config
	gwg    sync.WaitGroup
)

func start_chat_server(wg *sync.WaitGroup) {
	log.Println("INITIALIZING the server")

	chat_s = server.New()
	chat_s.Conf = conf

	err := chat_s.Serve()

	if err != nil {
		log.Fatalf("Could not listen -- addr: %s\n", conf.Server.Laddr)
		wg.Done()
	}
}


func start_admin_server(wg *sync.WaitGroup) {
	log.Printf("admin page -- Not Implemented Yet.")
	wg.Done()
}


func main() {
	conf = config.New()
	
	gwg.Add(2)

	go start_chat_server(&gwg)
	go start_admin_server(&gwg)

	gwg.Wait()
}
