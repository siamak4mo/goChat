package main

import (
	"encoding/json"
	"net/http"
	server "server/chat_server"
	"server/chat_server/serlog"
)

type AdminHandler struct {
	http.HandlerFunc `json:"-"`
	Info             string `json:"info"`
	Method           string `json:"usage"`
}

type AdminServer struct {
	Handlers     map[string]AdminHandler `json:"routes"`
	GoChatServer *server.Server          `json:"-"`
	Loger        *serlog.Log             `json:"-"`
}

func (s *AdminServer) Server() error {
	for p, fun := range s.Handlers {
		http.HandleFunc(p, fun.HandlerFunc)
	}

	admin_s.Loger.Pprintf("Listening on %s\n", conf.Admin.Addr)
	err := http.ListenAndServe(conf.Admin.Addr, nil)
	if err != nil {
		println("admin server -- could not listen on " + conf.Admin.Addr)
		return err
	}

	return nil
}

func NewAdminServer(server *server.Server) *AdminServer {
	h := make(map[string]AdminHandler)

	h["/"] = AdminHandler{
		HandlerFunc: root,
		Info:        "root page",
		Method:      "GET - no param",
	}
	h["/register"] = AdminHandler{
		HandlerFunc: reg_new_user,
		Info:        "register new user",
		Method:      "POST - username to sign up",
	}

	h["/chats/stat"] = AdminHandler{
		HandlerFunc: chat_stat,
		Info:        "statistics of chats",
		Method:      "GET - no param",
	}
	h["/chat/users"] = AdminHandler{
		HandlerFunc: chat_users,
		Info:        "show users of a chat room",
		Method:      "GET - URL params: ?chat=key (16 byte chat key)",
	}
	h["/chat/add"] = AdminHandler{
		HandlerFunc: chat_add,
		Info:        "make a new chat room",
		Method:      "POST - json req: {\"name\": \"chat name\", \"banner\": \"chat message of the day\"}",
	}
	h["/chat/remove"] = AdminHandler{
		HandlerFunc: chat_remove,
		Info:        "remove a chat room",
		Method:      "POST - json req: {\"chat key\": \"chat key (16 byte)\"}",
	}
	h["/users/stat"] = AdminHandler{
		HandlerFunc: user_stat,
		Info:        "show logged in users",
		Method:      "GET - no param",
	}
	h["/config/lookup"] = AdminHandler{
		HandlerFunc: config_lookup,
		Info:        "show the current server configuration",
		Method:      "GET - no param",
	}

	admin_s = &AdminServer{
		Handlers:     h,
		GoChatServer: server,
		Loger:        serlog.New(*conf, "Admin Server"),
	}

	return admin_s
}

func root(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		res := make(map[string]interface{})

		res["chat server"] = struct {
			Addr   string `json:"address"`
			Status string `json:"status"`
		}{
			Addr:   chat_s.Conf.Server.Addr,
			Status: "OK",
		}
		res["admin server"] = struct {
			Handlers map[string]AdminHandler `json:"Routes"`
			Name     string                  `json:"name"`
			Addr     string                  `json:"address"`
		}{
			Name:     "Admin Server",
			Addr:     conf.Admin.Addr,
			Handlers: admin_s.Handlers,
		}
		resp, err := json.Marshal(res)
		if err != nil {
			admin_s.Loger.Warnf("json.Marshal error: %s\n", err.Error())
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
			admin_s.Loger.Warnf("json.Marshal error: %s\n", err.Error())
		}
		w.Write(resp)
	}
}
func config_lookup(w http.ResponseWriter, r *http.Request) {
	admin_s.Loger.Warnf("Config Lookup Access\n")
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		resp, err := json.Marshal(chat_s.Conf)
		if err != nil {
			admin_s.Loger.Warnf("json.Marshal error: %s\n", err.Error())
		}
		w.Write(resp)
	}
}
func chat_stat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		res := make(map[string]interface{})

		for k, chat := range chat_s.Chats {
			res[k] = struct {
				Name  string `json:"name"`
				MOTD  string `json:"banner"`
				MEM_C int    `json:"member count"`
			}{
				Name:  chat.Name,
				MOTD:  chat.MOTD,
				MEM_C: len(chat.Members),
			}
		}
		resp, err := json.Marshal(res)
		if err != nil {
			admin_s.Loger.Warnf("json.Marshal error: %s\n", err.Error())
		}
		w.Write(resp)
	}
}
func chat_users(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		res := make(map[int]string)
		chat := r.URL.Query().Get("chat")

		if len(chat) == 16 && chat_s.Chats[chat] != nil {
			idx := 0
			for lp := range chat_s.Chats[chat].Members {
				res[idx] = lp.User.Username
				idx += 1
			}
		}
		resp, err := json.Marshal(res)
		if err != nil {
			admin_s.Loger.Warnf("json.Marshal error: %s\n", err.Error())
		}
		w.Write(resp)
	}
}
func chat_add(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		dec := json.NewDecoder(r.Body)
		data := struct {
			Name string `json:"name"`
			MOTD string `json:"banner"`
		}{}
		err := dec.Decode(&data)

		if err == nil {
			if !chat_s.HasChat(data.Name) {
				chat_s.AddNewChat(data.Name, data.MOTD)
				w.Write([]byte("Added\n"))
			}
		}
	}
}
func reg_new_user(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		dec := json.NewDecoder(r.Body)
		data := struct {
			Name string `json:"username"`
		}{}
		err := dec.Decode(&data)

		if err == nil {
			user_token := chat_s.RegisterUser(data.Name)
			w.Write([]byte("Token: " + user_token + "\n"))
		}else{
			w.Write([]byte("ops\n" + err.Error()))
		}
	}
}
func chat_remove(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		dec := json.NewDecoder(r.Body)
		data := struct {
			Key string `json:"chat key"`
		}{}
		err := dec.Decode(&data)

		if err == nil {
			if chat_s.HasChatKey(data.Key) {
				chat_s.RemoveChat(data.Key)
				w.Write([]byte("Removed\n"))
			}
		}
	}
}
