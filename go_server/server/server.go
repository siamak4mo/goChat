package server

import (
	"fmt"
	"log"
	"net"
	"server/server/config"
	"server/server/stoken"
	"strings"
	"unicode"
)

const (
	PAYLOAD_PADD    = 3
	MAXUSERNAME_LEN = 32

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
)

type User_t struct {
	Username  string
	Signature string
	ChatKey   string
}
type Packet struct {
	Conn    net.Conn
	Payload string
	User    User_t
	Type_t  Packet_t
}
type Chat struct {
	Members map[*Packet]bool
	Name    string
	MOTD    string
}
type Server struct {
	net.Listener
	Pac     chan Packet
	Clients map[string]*Packet
	Chats   map[string]*Chat
	Conf    config.Sconfig
}

func (u *User_t) String() string {
	return fmt.Sprintf("Username: %s\nSignature: %s\nChat: %s\n",
		u.Username, u.Signature, u.ChatKey)
}

func New() *Server {
	return &Server{
		Conf:    *config.New(),
		Chats:   make(map[string]*Chat),
		Clients: make(map[string]*Packet),
		Pac:     make(chan Packet),
	}
}

func (s *Server) HasChat(name string) bool {
	for k := range s.Chats {
		if strings.Compare(k, name) == 0 {
			return true
		}
	}
	return false
}

func newChat(name string, banner string) *Chat {
	return &Chat{
		Name:    name,
		MOTD:    banner,
		Members: make(map[*Packet]bool),
	}
}

func (s *Server) Serve() error {
	ln, err := net.Listen("tcp", s.Conf.Server.Laddr)
	if err != nil {
		return err
	}

	s.Listener = ln
	log.Printf("Listening on %s\n", s.Conf.Server.Laddr)

	// add two chates for testing
	s.Chats["echo"] = newChat("EcHo", "The Fundamental Chat!")
	s.Chats["69"] = newChat("69", "69 chat!")

	go s.handle_clients()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Could not accept a connection")
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
				if p.User == (User_t{}) {
					// disconnected before authentication
					log.Printf("ANONYMOUS DISCONNECTED")
				} else {
					log.Printf("%s disconnected\n", p.User.Username)
					if len(p.User.ChatKey) != 0 {
						delete(s.Chats[p.User.ChatKey].Members, s.Clients[p.Payload])
					}
				}
				delete(s.Clients, p.Payload)
			}
			break

		case P_login:
			tk := stoken.New_s(p.Payload, s.Conf) // exp payload: token  b64(username).signature
			if tk.Validate() {
				u := User_t{
					Username:  string(tk.Username),
					Signature: tk.Signature,
				}
				p.User = u
				if !s.username_exist(u.Username) {
					s.Clients[p.Conn.RemoteAddr().String()] = &p
					log.Printf("%s loged in\n", u.Username)
					p.Conn.Write([]byte("Loged in\n"))

				} else {
					p.Conn.Write([]byte("Already Loged in\n"))
				}
			} else {
				log.Printf("%s LOGIN FAILED\n", p.Payload)
				p.Conn.Write([]byte("Login Failed\n"))
				p.Conn.Close()
			}
			break

		case P_select_chat:
			if !s.HasChat(p.Payload) {
				p.Conn.Write([]byte("Chat doesn't exist\n"))
			} else {
				_lp := s.Clients[p.Conn.RemoteAddr().String()]
				if len(_lp.User.ChatKey) != 0 {
					delete(s.Chats[_lp.User.ChatKey].Members, _lp)
					_lp.User.ChatKey = p.Payload
					s.Chats[p.Payload].Members[_lp] = true
				} else {
					_lp.User.ChatKey = p.Payload
					s.Chats[p.Payload].Members[_lp] = true
				}

				p.Conn.Write([]byte("Welcome to " + p.Payload + "\n"))
				p.Conn.Write([]byte(s.Chats[p.Payload].MOTD + "\n"))
				go s.listen_client(p.Conn, _lp.User) // handle messages
			}
			break

		case P_new_text_message:
			for c := range s.Chats[p.User.ChatKey].Members {
				if strings.Compare(c.User.Username, p.User.Username) != 0 {
					c.Conn.Write([]byte(p.User.Username))
					c.Conn.Write([]byte("\n"))
					c.Conn.Write([]byte(p.Payload))
				}
			}
			break

		case P_signup:
			if !username_isvalid(p.Payload) {
				p.Conn.Write([]byte("Invalid username\n"))
				p.Conn.Close()
			} else {
				if !s.username_exist(p.Payload) {
					tk := stoken.New(s.Conf)
					tk.Username = []byte(p.Payload)
					tk.MkToken()

					p.Conn.Write([]byte("Token: " + tk.Token + "\n"))
					log.Printf("%s registered\n", p.Payload)
				} else {
					p.Conn.Write([]byte("User Already exists\n"))
				}
			}
			break

		case P_logout:
			delete(s.Clients, p.Payload) // exp payload: string(IP:PORT)
			p.Conn.Write([]byte("Loged out\n"))
			log.Printf("%s loged out", p.User.Username)

			go s.client_registry(p.Conn)
			break

		case P_whoami:
			p.Conn.Write([]byte(p.User.String()))
			p.Conn.Write([]byte("Addr: " + p.Conn.RemoteAddr().String() + "\n"))
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
		}
	}
}

func (s *Server) listen_client(conn net.Conn, u User_t) {
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
					Type_t: P_whoami,
					Conn:   conn,
					User:   u,
				}
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

// check username exists in loged in clients
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
