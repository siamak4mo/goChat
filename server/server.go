package main

import (
	"log"
	"net"
	"strings"
	"unicode"
)

const (
	LADDR      = "127.0.0.1"
	LPORT      = ":8080"
	LISTEN     = LADDR + LPORT
	LOGIN_KEYW = "LOGIN "
)

type Packet_t uint8

const (
	P_login Packet_t = iota + 1
	P_disconnected
	P_failed_login
	P_sign_up
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

		go reg_client(conn, p)
	}
}

func handle_clients(pac chan Packet) {
	clients := map[string]*Packet{}
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
				pac <- Packet{
					Type:    P_failed_login,
					Conn:    p.Conn,
					Payload: p.Conn.RemoteAddr().String(),
				}
				return
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

		case P_failed_login:
			log.Printf("%s LOGIN FAILED\n", p.Payload)
			p.Conn.Write([]byte("Login Failed\n"))
			p.Conn.Close()
			break

		case P_sign_up:
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

func reg_client(conn net.Conn, p chan Packet) {
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
		req := string(buffer[0 : n-1])

		if n > len(LOGIN_KEYW)+1 &&
			strings.Compare(req[0:len(LOGIN_KEYW)], LOGIN_KEYW) == 0 {
			p <- Packet{
				Type:    P_sign_up,
				Payload: req[len(LOGIN_KEYW) : n-1],
				Conn:    conn,
			}
		} else {
			p <- Packet{
				Type:    P_login,
				Conn:    conn,
				Payload: req,
			}
			return
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
		text := string(buffer[0:n])
		pac <- Packet{
			Type:    P_new_message,
			Conn:    conn,
			Payload: text,
			User:    u,
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
