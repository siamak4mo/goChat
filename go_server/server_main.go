package main

import (
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
	chat_s = server.New()
	chat_s.Conf = conf

	err := chat_s.Serve()

	if err != nil {
		wg.Done()
	}
}


func start_admin_server(wg *sync.WaitGroup) {
	admin_s = NewAdminServer(chat_s)
	err := admin_s.Server()

	if err != nil {
		wg.Done()
	}
}

func main() {
	conf = config.New()
	
	gwg.Add(2)

	go start_chat_server(&gwg)
	go start_admin_server(&gwg)

	gwg.Wait()
}
