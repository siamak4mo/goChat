package main

import (
	"log"
	"net"
	"strings"
)

const (
	LADDR = "127.0.0.1"
	LPORT = ":8080"
	LISTEN = LADDR+LPORT
)

type Packet_t uint8
const (
	P_loged_in Packet_t = iota + 1
	P_disconnected
	P_failed_login
	P_login_req
	P_new_message
)

type User_t struct{
	Username string
	Signature string
}
type Packet struct{
	Type Packet_t
	Conn net.Conn
	User User_t
	Payload string
}


func main(){
	log.Println("INITIALIZING server")

	ln, err := net.Listen("tcp", LISTEN)
	if err!=nil{
		log.Fatalf("Could open port%s\n", LPORT)
	}
	log.Printf("Listening on %s\n", LISTEN)

	p := make(chan Packet)
	
	go handle_clients(p)
	
	for{
		conn, err := ln.Accept()
		if err!=nil{
			log.Println("Could not accept a connection")
			continue
		}

		go reg_client(conn, p)	
	}
}


func handle_clients(pac chan Packet){
	clients := map[string]*Packet{}
	for{
		p := <- pac

		switch p.Type{
			case P_disconnected:
			if len(p.Payload) != 0 {
				log.Printf("%s DISCONNECTED\n", p.Payload)
				delete(clients, p.Payload)
			}
			break

			case P_loged_in:
			clients[p.Payload] = &p
			log.Printf("%s CONNECTED\n", p.User.Username)
			p.Conn.Write([]byte("loged in\n"))
			break

			case P_new_message:
			for _, c := range clients{
				if c.Conn != p.Conn{
					c.Conn.Write([]byte(c.User.Username))
					c.Conn.Write([]byte(p.Payload))
				}
			}
			break

			case P_failed_login:
			log.Printf("%s LOGIN FAILED\n", p.Payload)
			break

			case P_login_req:
			_exists := false
			for _,c := range clients{
				if c.User.Username == p.Payload{
					_exists = true
					break
				}
			}
			if !_exists{
				tk := Init_token()
				tk.Username = []byte(p.Payload)
				tk.MkToken()
				
				p.Conn.Write([]byte(tk.Token))
				p.Conn.Close()
				log.Printf("%s just registered", p.Payload)
			}else{
				p.Conn.Write([]byte("User Already exists"))
				p.Conn.Close()
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
				Type: P_disconnected,
				Conn: conn,
				Payload: conn.RemoteAddr().String(),
			}
			conn.Close();
			return
		}
		token := string(buffer[0:n-1])
		
		if n>7 && strings.Compare(token[0:6], "LOGIN ") == 0{
			p <- Packet{
				Type: P_login_req,
				Payload: token[6:n-1],
				Conn: conn,
			}
		}else{
			tk := Init_stoken(token)
			if tk.Validate(){
				p <- Packet{
					Type: P_loged_in,
					Payload: token,
					User: User_t{
						Username: string(tk.Username),
						Signature: tk.Signature,
					},
					Conn: conn,
				}
				listen_client(conn, p)
			}else{
				conn.Close();
				p <- Packet{
					Type: P_failed_login,
					Conn: conn,
					Payload: conn.RemoteAddr().String(),
				}
				return
			}
		}
	}
}

func listen_client(conn net.Conn, p chan Packet){
	buffer := make([]byte, 64)
	for{
		n, err := conn.Read(buffer)
		if err != nil {
			conn.Close();
			p <- Packet{
				Type: P_disconnected,
				Conn: conn,
			}
			return
		}
		text := string(buffer[0:n])
		p <- Packet{
			Type: P_new_message,
			Conn: conn,
			Payload: text,
		}
	}
}
