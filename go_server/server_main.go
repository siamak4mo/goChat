package main

import (
	"os"
	server "server/chat_server"
	"server/chat_server/config"
	"strings"
	"sync"
)

var (
	chat_s  *server.Server
	admin_s *AdminServer
	conf    *config.Config
	gwg     sync.WaitGroup
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
	if len(os.Args) == 3 &&
		strings.Compare(os.Args[1], "-C") == 0 {
		conf = config.New(os.Args[2])
	} else {
		conf = config.New()
	}

	gwg.Add(2)

	go start_chat_server(&gwg)
	go start_admin_server(&gwg)

	gwg.Wait()
}
