package main

import (
	"log"
	"net"
	"unicode"
)

const (
	LADDR  = "127.0.0.1"
	LPORT  = ":8080"
	LISTEN = LADDR + LPORT
)

type Packet_t uint8

const (
	P_login Packet_t = iota + 1
	P_signup
	P_disconnected
	P_new_message
)

type User_t struct {
	Username  string
	Signature string
}
type Packet struct {
	Conn    net.Conn
	Payload string
	User    User_t
	Type    Packet_t
}

// map from client "IP:PORT" -> login packet
var clients = map[string]*Packet{}

func main() {
	log.Println("INITIALIZING server")

	ln, err := net.Listen("tcp", LISTEN)
	if err != nil {
		log.Fatalf("Could open port%s\n", LPORT)
	}
	log.Printf("Listening on %s\n", LISTEN)

	p := make(chan Packet)

	go handle_clients(p)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Could not accept a connection")
			continue
		}

		go client_registry(conn, p)
	}
}

func handle_clients(pac chan Packet) {
	for {
		p := <-pac

		switch p.Type {
		case P_disconnected:
			if len(p.Payload) != 0 {
				log.Printf("%s DISCONNECTED\n", p.Payload)
				delete(clients, p.Payload)
			}
			break

		case P_login:
			tk := Init_stoken(p.Payload)
			if tk.Validate() {
				u := User_t{
					Username:  string(tk.Username),
					Signature: tk.Signature,
				}
				clients[p.Conn.RemoteAddr().String()] = &p
				log.Printf("%s CONNECTED\n", u.Username)
				p.Conn.Write([]byte("Loged in\n"))
				go listen_client(p.Conn, pac, u)
			} else {
				log.Printf("%s LOGIN FAILED\n", p.Payload)
				p.Conn.Write([]byte("Login Failed\n"))
				p.Conn.Close()
			}
			break

		case P_new_message:
			for _, c := range clients {
				if c.Conn != p.Conn {
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
				_exists := false
				for _, c := range clients {
					if c.User.Username == p.Payload {
						_exists = true
						break
					}
				}
				if !_exists {
					tk := Init_token()
					tk.Username = []byte(p.Payload)
					tk.MkToken()

					p.Conn.Write([]byte(tk.Token))
					p.Conn.Close()
					log.Printf("%s just registered\n", p.Payload)
				} else {
					p.Conn.Write([]byte("User Already exists\n"))
					p.Conn.Close()
				}
			}
			break
		}
	}
}

func client_registry(conn net.Conn, p chan Packet) {
	buffer := make([]byte, 256)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			p <- Packet{
				Type:    P_disconnected,
				Conn:    conn,
				Payload: conn.RemoteAddr().String(),
			}
			conn.Close()
			return
		}

		if n > 3 {
			switch buffer[0] {
			case 'L': // login
				p <- Packet{
					Type:    P_login,
					Conn:    conn,
					Payload: string(buffer[2 : n-1]),
				}
				return

			case 'S': // signup
				p <- Packet{
					Type:    P_signup,
					Conn:    conn,
					Payload: string(buffer[2 : n-1]),
				}
				return
			}
		}
	}
}

func listen_client(conn net.Conn, pac chan Packet, u User_t) {
	buffer := make([]byte, 64)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			pac <- Packet{
				Type:    P_disconnected,
				Conn:    conn,
				Payload: conn.RemoteAddr().String(),
			}
			return
		}

		if n > 3 {
			switch buffer[0] {
			case 'T': // text message
				pac <- Packet{
					Type:    P_new_message,
					Conn:    conn,
					Payload: string(buffer[2:n]),
					User:    u,
				}
				break
			}
		}
	}
}

func username_isvalid(name string) bool {
	if len(name) > 32 || len(name) == 0 {
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
