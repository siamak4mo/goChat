package main

import (
	"encoding/json"
	"net/http"
	server "server/chat_server"
	"server/config"
	"server/logger"
)

type routeHandler struct {
	http.HandlerFunc `json:"-"`
	Info             string `json:"info"`
	Method           string `json:"usage"`
}

type AdminServer struct {
	Handlers   map[string]routeHandler `json:"routes"`
	ChatServer *server.Server          `json:"-"`
	Loger      *logger.Log             `json:"-"`
	Conf       *config.Config          `json:"-"`
}

func (s *AdminServer) Serve() error {
	for p, fun := range s.Handlers {
		http.HandleFunc(p, fun.HandlerFunc)
	}

	s.Loger.Pprintf("Listening on %s\n", s.Conf.Admin.Addr)
	err := http.ListenAndServe(s.Conf.Admin.Addr, nil)
	if err != nil {
		println("admin server -- could not listen on " + s.Conf.Admin.Addr)
		return err
	}

	return nil
}

func NewAdminServer(controller *main_controller) *AdminServer {
	h := make(map[string]routeHandler)

	s := &AdminServer{
		Handlers:   h,
		ChatServer: controller.chat_s,
		Loger:      logger.New(*controller.config, "Admin Server"),
		Conf:       controller.config,
	}

	h["/"] = routeHandler{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			s.root(w, r)
		},
		Info:   "root page",
		Method: "GET - no param",
	}
	h["/register"] = routeHandler{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			s.reg_new_user(w, r)
		},
		Info:   "register new user",
		Method: "POST - username to sign up",
	}

	h["/chats/stat"] = routeHandler{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			s.chat_stat(w, r)
		},
		Info:   "statistics of chats",
		Method: "GET - no param",
	}
	h["/chat/users"] = routeHandler{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			s.chat_users(w, r)
		},
		Info:   "show users of a chat room",
		Method: "GET - URL params: ?chat=key (16 byte chat key)",
	}
	h["/chat/add"] = routeHandler{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			s.chat_add(w, r)
		},
		Info:   "make a new chat room",
		Method: "POST - json req: {\"name\": \"chat name\", \"banner\": \"chat message of the day\"}",
	}
	h["/chat/remove"] = routeHandler{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			s.chat_remove(w, r)
		},
		Info:   "remove a chat room",
		Method: "POST - json req: {\"chat key\": \"chat key (16 byte)\"}",
	}
	h["/users/stat"] = routeHandler{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			s.user_stat(w, r)
		},
		Info:   "show logged in users",
		Method: "GET - no param",
	}
	h["/config/lookup"] = routeHandler{
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			s.config_lookup(w, r)
		},
		Info:   "show the current server configuration",
		Method: "GET - no param",
	}

	return s
}

func (s *AdminServer) root(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		res := make(map[string]interface{})

		res["chat server"] = struct {
			Addr   string `json:"address"`
			Status string `json:"status"`
		}{
			Addr:   s.ChatServer.Conf.Server.Addr,
			Status: "OK",
		}
		res["admin server"] = struct {
			Handlers map[string]routeHandler `json:"Routes"`
			Name     string                  `json:"name"`
			Addr     string                  `json:"address"`
		}{
			Name:     "Admin Server",
			Addr:     s.Conf.Admin.Addr,
			Handlers: s.Handlers,
		}
		resp, err := json.Marshal(res)
		if err != nil {
			s.Loger.Warnf()("json.Marshal error: %s\n", err.Error())
		}
		w.Write(resp)
	}
}
func (s *AdminServer) user_stat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		res := make(map[string]interface{})

		for addr, lp := range s.ChatServer.Clients {
			res[lp.User.Username] = struct {
				Addr     string `json:"address"`
				Token    string `json:"token"`
				ChatKey  string `json:"chat key"`
				ChatName string `json:"chat name"`
			}{
				Addr:     addr,
				Token:    lp.Payload,
				ChatKey:  lp.User.ChatKey,
				ChatName: s.ChatServer.ChatName(*lp),
			}
		}
		resp, err := json.Marshal(res)
		if err != nil {
			s.Loger.Warnf()("json.Marshal error: %s\n", err.Error())
		}
		w.Write(resp)
	}
}
func (s *AdminServer) config_lookup(w http.ResponseWriter, r *http.Request) {
	s.Loger.Warnf()("Config Lookup Access\n")
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		resp, err := json.Marshal(s.Conf)
		if err != nil {
			s.Loger.Warnf()("json.Marshal error: %s\n", err.Error())
		}
		w.Write(resp)
	}
}
func (s *AdminServer) chat_stat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		res := make(map[string]interface{})

		for k, chat := range s.ChatServer.Chats {
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
			s.Loger.Warnf()("json.Marshal error: %s\n", err.Error())
		}
		w.Write(resp)
	}
}
func (s *AdminServer) chat_users(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		res := make(map[int]string)
		chat := r.URL.Query().Get("chat")

		if len(chat) == 16 && s.ChatServer.Chats[chat] != nil {
			idx := 0
			for lp := range s.ChatServer.Chats[chat].Members {
				res[idx] = lp.User.Username
				idx += 1
			}
		}
		resp, err := json.Marshal(res)
		if err != nil {
			s.Loger.Warnf()("json.Marshal error: %s\n", err.Error())
		}
		w.Write(resp)
	}
}
func (s *AdminServer) chat_add(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		dec := json.NewDecoder(r.Body)
		data := struct {
			Name string `json:"name"`
			MOTD string `json:"banner"`
		}{}
		err := dec.Decode(&data)

		if err == nil {
			if !s.ChatServer.HasChat(data.Name) {
				s.ChatServer.AddNewChat(data.Name, data.MOTD)
				w.Write([]byte("Added\n"))
			}
		}
	}
}
func (s *AdminServer) reg_new_user(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		dec := json.NewDecoder(r.Body)
		data := struct {
			Name string `json:"username"`
		}{}
		err := dec.Decode(&data)

		if err == nil {
			user_token := s.ChatServer.RegisterUser(data.Name)
			w.Write([]byte("Token: " + user_token + "\n"))
		} else {
			w.Write([]byte("ops\n" + err.Error()))
		}
	}
}
func (s *AdminServer) chat_remove(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		dec := json.NewDecoder(r.Body)
		data := struct {
			Key string `json:"chat key"`
		}{}
		err := dec.Decode(&data)

		if err == nil {
			if s.ChatServer.HasChatKey(data.Key) {
				s.ChatServer.RemoveChat(data.Key)
				w.Write([]byte("Removed\n"))
			}
		}
	}
}
