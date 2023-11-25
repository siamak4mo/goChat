package main

import (
	"encoding/json"
	"net/http"
	"server/server"
)

var (
	admin_s *AdminServer
)

type AdminHandler struct {
	http.HandlerFunc `json:"-"`
	Info             string `json:"Info"`
}

type AdminServer struct {
	Handlers     map[string]AdminHandler `json:"Routes"`
	GoChatServer *server.Server          `json:"-"`
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
	h["/users/stat"] = AdminHandler{
		HandlerFunc: user_stat,
		Info:        "show loged in users",
	}
	h["/config/lookup"] = AdminHandler{
		HandlerFunc: config_lookup,
		Info:        "show the current server configuration",
	}

	admin_s = &AdminServer{
		Handlers:     h,
		GoChatServer: server,
	}

	return admin_s
}

func root(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		resp, err := json.Marshal(admin_s)
		if err != nil {
			println(err.Error())
		}
		w.Write(resp)
	}
}
func user_stat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		res := make(map[string]interface{})

		for addr, lp := range chat_s.Clients {
			res[lp.User.Username] = struct {
				Addr     string `json:"address"`
				Token    string `json:"token"`
				ChatKey  string `json:"chat key"`
				ChatName string `json:"chat name"`
			}{
				Addr:     addr,
				Token:    lp.Payload,
				ChatKey:  lp.User.ChatKey,
				ChatName: chat_s.ChatName(*lp),
			}
		}
		resp, err := json.Marshal(res)
		if err != nil {
			println(err.Error())
		}
		w.Write(resp)
	}
}
func config_lookup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		resp, err := json.Marshal(chat_s.Conf)
		if err != nil {
			println(err.Error())
		}
		w.Write(resp)
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
