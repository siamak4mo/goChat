package main

import (
	"os"
	server "server/chat_server"
	"server/chat_server/config"
	"strings"
	"sync"
)

type main_controller struct {
	chat_s  *server.Server
	admin_s *AdminServer
	config  *config.Config
	gwg     sync.WaitGroup
}

func (mc *main_controller) start_chat_server() {
	err := mc.chat_s.Serve()

	if err != nil {
		mc.gwg.Done()
	}
}

func (mc *main_controller) start_admin_server() {
	err := mc.admin_s.Server()

	if err != nil {
		mc.gwg.Done()
	}
}

func main() {
	controller := main_controller{}

	if len(os.Args) == 3 &&
		strings.Compare(os.Args[1], "-C") == 0 {
		controller.config = config.New(os.Args[2])
	} else {
		controller.config = config.New()
	}

	/* initialize the chat server and the admin server */
	controller.chat_s = server.New()
	controller.chat_s.Conf = controller.config
	/* ALWAYS after initialization of the chat server */
	controller.admin_s = NewAdminServer(&controller)

	controller.gwg.Add(2)
	go controller.start_chat_server()
	go controller.start_admin_server()

	controller.gwg.Wait()
}
