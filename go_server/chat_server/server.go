package server

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"server/config"
	"server/serlog"
	"server/chat_server/stoken"
	"strings"
	"unicode"
)

const (
	PAYLOAD_PADD    = 3
	MAXUSERNAME_LEN = 32
	CHATID_LEN      = 16

	S_BUFF_S = 64
	M_BUFF_S = 256
	B_BUFF_S = 512
)

type Packet_t uint8

const (
	P_login Packet_t = iota + 1
	P_select_chat
	P_signup
	P_logout
	P_disconnected
	P_new_text_message
	P_whoami
	P_list_chats
)

type User_t struct {
	Username  string
	Signature string
	ChatKey   string
}
type Packet struct {
	User    *User_t
	Conn    net.Conn
	Payload string
	Type_t  Packet_t
}
type Chat struct {
	Members map[*Packet]bool // set of login packets
	ChatKey string
	Name    string
	MOTD    string // message of the day
}
type Server struct {
	net.Listener
	Pac     chan Packet
	Clients map[string]*Packet // map from "ip:port" to login packet
	Chats   map[string]*Chat   // map from ChatKey to Chat
	Conf    *config.Config
	Log     *serlog.Log
}

func (u *User_t) String() string {
	return fmt.Sprintf("Username: %s\nSignature: %s\n",
		u.Username, u.Signature)
}

func New() *Server {
	return &Server{
		Chats:   make(map[string]*Chat),
		Clients: make(map[string]*Packet),
		Pac:     make(chan Packet),
	}
}

func (s *Server) HasChatKey(key string) bool {
	for k := range s.Chats {
		if strings.Compare(k, key) == 0 {
			return true
		}
	}
	return false
}

func (s *Server) HasChat(name string) bool {
	for _, c := range s.Chats {
		if strings.Compare(c.Name, name) == 0 {
			return true
		}
	}
	return false
}

func (s Server) ChatName(p Packet) string {
	res := s.Chats[p.User.ChatKey]
	if res == nil || len(res.Name) == 0 {
		return "Internal Server Error"
	} else {
		return res.Name
	}
}

func (p *Packet) RemoteAddr() string {
	return p.Conn.RemoteAddr().String()
}

func (p *Packet) Swrite(data string, s *Server) {
	_, err := io.WriteString(p.Conn, data)

	if err != nil {
		s.Pac <- Packet{
			Type_t:  P_disconnected,
			Conn:    p.Conn,
			User:    p.User,
			Payload: p.RemoteAddr(),
		}
	}
}

func newChat(name string, banner string) *Chat {
	return &Chat{
		ChatKey: hex.EncodeToString(sha1.New().Sum([]byte(name)))[0:CHATID_LEN],
		Name:    name,
		MOTD:    banner,
		Members: make(map[*Packet]bool),
	}
}

func (s *Server) AddNewChat(name string, banner string) {
	c := newChat(name, banner)
	s.Log.Infof("chat %s added\n", c.ChatKey)
	s.Chats[c.ChatKey] = c
}

func (s *Server) RemoveChat(key string) {
	s.Log.Infof("chat %s removed\n", key)
	for lp := range s.Chats[key].Members {
		lp.Swrite("this chat no longer exists, login to another chat\n", s)
		lp.User.ChatKey = ""
	}
	delete(s.Chats, key)
}

func (s *Server) RegisterUser(name string) string {
	tk := stoken.New(s.Conf)
	tk.Username = []byte(name)
	tk.MkToken()

	s.Log.Infof("%s registered\n", tk.Signature[0:16])

	return tk.Token
}

func (s *Server) Serve() error {
	s.Log = serlog.New(*s.Conf, "Chat Server ")
	ln, err := net.Listen("tcp", s.Conf.Server.Addr)

	if err != nil {
		s.Log.Panicf("Could not listen on %v\n",
			serlog.Nop, s.Conf.Server.Addr)
		return err
	}

	s.Listener = ln
	s.Log.Pprintf("Listening on %s\n", s.Conf.Server.Addr)

	for i, name := range s.Conf.Server.Chats {
		s.AddNewChat(name, s.Conf.Server.ChatMOTD[i])
	}

	go s.handle_clients()

	for {
		conn, err := ln.Accept()
		if err != nil {
			s.Log.Warnf("could not accept a connection\n")
			continue
		}

		go s.client_registry(conn)
	}
}

// main logic
func (s *Server) handle_clients() {
	for {
		p := <-s.Pac

		switch p.Type_t {
		case P_disconnected:
			if len(p.Payload) != 0 { // exp payload: string(IP:PORT)
				_u := &User_t{}
				if p.User != nil {
					_u = p.User
					delete(s.Clients, p.Payload)
					s.Log.Debugf("%s disconnected\n", _u.Username)
				} else if s.Clients[p.Payload] != nil {
					_u = s.Clients[p.Payload].User
					delete(s.Clients, p.Payload)
					s.Log.Debugf("%s disconnected\n", _u.Username)
				} else {
					s.Log.Debugf("ANONYMOUS DISCONNECTED\n")
				}

				if len(_u.ChatKey) != 0 {
					delete(s.Chats[_u.ChatKey].Members, s.Clients[p.Payload])
				}
			}
			break

		case P_login:
			tk := stoken.New_s(p.Payload, s.Conf) // exp payload: token  b64(username).signature
			if tk.Validate() {
				u := &User_t{
					Username:  string(tk.Username),
					Signature: tk.Signature,
				}
				p.User = u
				if !s.username_exist(u.Username) {
					s.Clients[p.RemoteAddr()] = &p
					s.Log.Debugf("%s logged in\n", u.Username)
					p.Swrite("Logged in\n", s)

				} else {
					p.Swrite("Already Logged in\n", s)
				}
			} else {
				s.Log.Infof("LOGIN FAILED\n")
				p.Swrite("Login Failed\n", s)
				p.Conn.Close()
			}
			break

		case P_select_chat:
			p_login := s.Clients[p.RemoteAddr()]
			if p_login == nil {
				p.Swrite("you are not logged in\n", s)
				go s.client_registry(p.Conn)
				break
			}
			if !s.HasChatKey(p.Payload) {
				p.Swrite("Chat doesn't exist\n", s)
				go s.listen_client(p.Conn, p_login.User) // handle messages
			} else {
				if len(p_login.User.ChatKey) != 0 {
					delete(s.Chats[p_login.User.ChatKey].Members, p_login)
					p_login.User.ChatKey = p.Payload
					s.Chats[p.Payload].Members[p_login] = true
				} else {
					p_login.User.ChatKey = p.Payload
					s.Chats[p.Payload].Members[p_login] = true
				}

				p.Swrite(s.Chats[p.Payload].MOTD+"\n", s)
				go s.listen_client(p.Conn, p_login.User) // handle messages
			}
			break

		case P_new_text_message:
			if len(p.User.ChatKey) == 0 {
				break
			}
			for c := range s.Chats[p.User.ChatKey].Members {
				if strings.Compare(c.User.Username, p.User.Username) != 0 {
					go c.Swrite(p.User.Username+"\n"+p.Payload, s)
				}
			}
			break

		case P_signup:
			if !username_isvalid(p.Payload) {
				go func() {
					p.Swrite("Invalid username\n", s)
					p.Conn.Close()
				}()
			} else {
				// in this level we only make not trusted users
				// so they must begin  with `NT_`
				if len(p.Payload) < 3 ||
					strings.Compare(p.Payload[0:3], "NT_") != 0 {
					p.Swrite("Invalid username, only `NT_username` is allowed\n", s)
					p.Conn.Close()
					break
				}
				if !s.username_exist(p.Payload) {
					user_token := s.RegisterUser(p.Payload)
					p.Swrite("Token: "+user_token+"\n", s)
				} else {
					p.Swrite("User Already exists\n", s)
				}
			}
			break

		case P_logout:
			_u := &User_t{}
			if p.User != nil {
				_u = p.User
				s.Log.Debugf("%s logged out\n", _u.Username)
				p.Swrite("Logged out\n", s)
			} else {
				if s.Clients[p.Payload] != nil {
					_u = s.Clients[p.Payload].User
					s.Log.Debugf("%s logged out\n", _u.Username)
					p.Swrite("Logged out\n", s)
				} else {
					p.Swrite("you are not logged in\n", s)
				}
			}

			if len(_u.ChatKey) != 0 {
				delete(s.Chats[_u.ChatKey].Members, s.Clients[p.Payload])
			}
			delete(s.Clients, p.Payload) // exp payload: string(IP:PORT)

			go s.client_registry(p.Conn)
			break

		case P_whoami: // exp payload: string(IP:PORT)
			_u := &User_t{}
			if p.User != nil {
				_u = p.User
			} else {
				if s.Clients[p.Payload] != nil {
					_u = s.Clients[p.Payload].User
				} else {
					p.Swrite("Anonymous\n", s)
					break
				}
			}

			if len(_u.ChatKey) != 0 {
				p.Swrite(fmt.Sprintf("%sChat: %s\nAddr: %s\n",
					_u.String(), s.ChatName(p), p.Payload), s)
			} else {
				p.Swrite(fmt.Sprintf("%sAddr: %s\n",
					_u.String(), p.Payload), s)
			}
			break

		case P_list_chats:
			for k, v := range s.Chats {
				p.Swrite(fmt.Sprintf("ChatID: %s -- Name: %s\n", k, v.Name), s)
			}
			p.Swrite("EOF\n", s)
			break
		}
	}
}

func (s *Server) client_registry(conn net.Conn) {
	buffer := make([]byte, M_BUFF_S)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			s.Pac <- Packet{
				Type_t:  P_disconnected,
				Conn:    conn,
				Payload: conn.RemoteAddr().String(),
			}
			conn.Close()
			return
		}

		if n > PAYLOAD_PADD {
			switch buffer[0] {
			case 'L': // login
				s.Pac <- Packet{
					Type_t:  P_login,
					Conn:    conn,
					Payload: string(buffer[PAYLOAD_PADD-1 : n-1]),
				}
				break

			case 'S': // signup
				s.Pac <- Packet{
					Type_t:  P_signup,
					Conn:    conn,
					Payload: string(buffer[PAYLOAD_PADD-1 : n-1]),
				}
				break

			case 'C': // select chat
				s.Pac <- Packet{
					Type_t:  P_select_chat,
					Conn:    conn,
					Payload: string(buffer[PAYLOAD_PADD-1 : n-1]),
				}
				return
			}
		} else if n > 0 {
			switch buffer[0] {
			case 'C': // list chats
				s.Pac <- Packet{
					Type_t: P_list_chats,
					Conn:   conn,
				}
				break

			case 'L': // to log out
				s.Pac <- Packet{
					Type_t:  P_logout,
					Conn:    conn,
					Payload: conn.RemoteAddr().String(),
				}
				break

			case 'W': // whoami
				s.Pac <- Packet{
					Type_t:  P_whoami,
					Conn:    conn,
					Payload: conn.RemoteAddr().String(),
				}
			}
		}
	}
}

func (s *Server) listen_client(conn net.Conn, u *User_t) {
	buffer := make([]byte, B_BUFF_S)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			// user has disconnected
			s.Pac <- Packet{
				Type_t:  P_disconnected,
				Conn:    conn,
				Payload: conn.RemoteAddr().String(),
				User:    u,
			}
			return
		}

		if n > PAYLOAD_PADD {
			switch buffer[0] {
			case 'T':
				// got text message
				s.Pac <- Packet{
					Type_t:  P_new_text_message,
					Conn:    conn,
					Payload: string(buffer[PAYLOAD_PADD-1 : n]),
					User:    u,
				}
				break

			case 'C':
				s.Pac <- Packet{
					Type_t:  P_select_chat,
					Conn:    conn,
					Payload: string(buffer[PAYLOAD_PADD-1 : n-1]),
				}
				return
			}
		} else if n > 0 {
			switch buffer[0] {
			case 'L':
				// to logout
				s.Pac <- Packet{
					Type_t:  P_logout,
					Conn:    conn,
					Payload: conn.RemoteAddr().String(),
					User:    u,
				}
				return

			case 'W':
				// send whoami
				s.Pac <- Packet{
					Type_t:  P_whoami,
					Conn:    conn,
					Payload: conn.RemoteAddr().String(),
					User:    u,
				}

			case 'C':
				s.Pac <- Packet{
					Type_t: P_list_chats,
					Conn:   conn,
				}
				break
			}
		}
	}
}

func username_isvalid(name string) bool {
	if len(name) > MAXUSERNAME_LEN || len(name) == 0 {
		return false
	}
	for i := 0; i < len(name); i++ {
		if name[i] > unicode.MaxASCII ||
			name[i] == '\n' || name[i] == '\r' {
			return false
		}
	}
	return true
}

// check username exists in logged in clients
func (s *Server) username_exist(uname string) bool {
	if len(uname) == 0 {
		return true // to prevent null user addition
	}
	for _, login_pac := range s.Clients {
		if strings.Compare(login_pac.User.Username, uname) == 0 {
			return true
		}
	}

	return false
}
