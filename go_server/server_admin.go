package main

import (
	"io"
	"net/http"
	"server/server"
)

var (
	admin_s *AdminServer
)

type AdminHandler struct {
	http.HandlerFunc
	Info string
}

type AdminServer struct {
	Handlers     map[string]AdminHandler
	GoChatServer *server.Server
}

func (s *AdminServer) Server() error {
	for p, fun := range s.Handlers {
		http.HandleFunc(p, fun.HandlerFunc)
	}

	err := http.ListenAndServe(":4242", nil)
	if err != nil {
		println("admin server -- could not listen on :4242")
		return err
	}

	return nil
}

func NewAdminServer(server *server.Server) *AdminServer {
	h := make(map[string]AdminHandler)

	h["/"] = AdminHandler{
		HandlerFunc: root,
		Info:        "GET / PAGE",
	}
	h["/chats/stat"] = AdminHandler{
		HandlerFunc: chat_stat,
		Info:        "statistics of chats",
	}
	h["/chat/add"] = AdminHandler{
		HandlerFunc: chat_add,
		Info:        "make a new chat room",
	}
	h["/chat/remove"] = AdminHandler{
		HandlerFunc: chat_remove,
		Info:        "remove a chat room",
	}

	admin_s = &AdminServer{
		Handlers:     h,
		GoChatServer: server,
	}

	return admin_s
}

func root(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		for k, v := range admin_s.Handlers {
			io.WriteString(w, "Route: "+k+"\nInfo: "+v.Info+"\n\n")
		}
	}
}
func chat_stat(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.Method))
}
func chat_add(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.Method))
}
func chat_remove(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(r.Method))
}
