package server

import (
	"fmt"
	"io"
	"net"
	server "server/chat_server"
	"server/chat_server/config"
	"strings"
	"testing"
	"time"
)

var (
	chat_s *server.Server
	conf   *config.Config
)

var (
	csc1 *ChatSerConn
	csc2 *ChatSerConn
)

type chatCommand uint8

const (
	C_signup chatCommand = iota + 1
	C_login_out
	C_text
	C_chat_select
)

type ChatSerConn struct {
	net.Conn
	Token    string
	Username string
	MaxRetry int
}

func NewCSC() *ChatSerConn {
	return &ChatSerConn{
		MaxRetry: 4,
	}
}

func (csc *ChatSerConn) mkConn() error {
	estab := false
	var err error
	for csc.MaxRetry != 0 {
		c, err := net.Dial("tcp", ":8080") // default port
		if err == nil {
			csc.Conn = c
			estab = true
			break
		} else {
			csc.MaxRetry -= 1
			fmt.Printf("retrying to connect to the chat server -- left %v times\n",
				csc.MaxRetry)
		}

		time.Sleep(1 * time.Second)
	}

	if !estab {
		return err
	} else {
		return nil
	}
}

func init() {
	conf = config.Default()
	chat_s = server.New()
	chat_s.Conf = conf

	go func() {
		defer println("fatal -- unreachable")
		if e := chat_s.Serve(); e != nil {
			println("fatal -- " + e.Error())
		}
	}()

	csc1 = NewCSC()
	csc2 = NewCSC()
}

func (csc *ChatSerConn) send2chat(command chatCommand, mess string) {
	var comm string
	switch command {
	case C_signup:
		comm = "S "
		break

	case C_login_out:
		comm = "L "
		break

	case C_text:
		comm = "T "
		break

	case C_chat_select:
		comm = "C "
		break
	}

	io.WriteString(csc.Conn, comm+mess+"\n")
}

func (csc *ChatSerConn) readFchat() string {
	buf := make([]byte, 512)
	n, err := csc.Read(buf)

	if err != nil {
		return err.Error()
	}

	return string(buf[0 : n-1]) // skip `\n` at the end
}

func Test_chatserver_is_up(t *testing.T) {
	e := csc1.mkConn()
	if e != nil {
		t.Fatalf("cannot connect to the chat server.")
	}
}

func Test_Signup_Login(t *testing.T) {
	csc1.Username = "my name"
	csc1.send2chat(C_signup, csc1.Username) // send signup request
	res := csc1.readFchat()

	if strings.Compare(res,
		"Token: bXkgbmFtZQ==.2548acd8b40019cffd702fcf87ba50bfc8c948d3247894c7a89c5fcc847c21ff") != 0 {
		t.Fatalf("Wrong Token")
	}
	csc1.Token = res[7:] // `Token: xxx`[7:] = `xxx`

	csc1.send2chat(C_login_out, csc1.Token) // send login request

	if strings.Compare(csc1.readFchat(), "Loged in") != 0 {
		t.Fatalf("login failed")
	}
}

func Test_Second_Conn(t *testing.T) {
	e := csc2.mkConn()
	if e != nil {
		t.Fatalf("the chat server does not handle second connection.")
	}
}

func Test_Second_Login(t *testing.T) {
	csc2.send2chat(C_login_out, csc1.Token) // login with loged in username

	if strings.Compare(csc2.readFchat(), "Already Loged in") != 0 {
		t.Fatalf("double login.")
	}

	csc2.Username = "user-2-"
	csc2.send2chat(C_signup, csc2.Username)
	_tmp := csc2.readFchat()

	if strings.Compare(_tmp[0:6], "Token:") != 0 {
		t.Fatalf("second login lailed.")
	}

	csc2.Token = _tmp[7:]

	csc2.send2chat(C_login_out, csc2.Token)

	if strings.Compare(csc2.readFchat(), "Loged in") != 0 {
		t.Fatalf("login failed")
	}
}

func Test_Select_Chat_Messaging(t *testing.T) {
	chat1 := "4563486fda39a3ee"
	chat2 := "48434bda39a3ee5e"

	csc1.send2chat(C_chat_select, chat1)
	if strings.Compare(csc1.readFchat(), "Chat doesn't exist") == 0 {
		t.Fatalf("chat %s does not exist", chat1)
	}
	csc2.send2chat(C_chat_select, chat1)
	if strings.Compare(csc2.readFchat(), "Chat doesn't exist") == 0 {
		t.Fatalf("chat %s does not exist.", chat1)
	}

	mess := "Hi"
	csc1.send2chat(C_text, mess)
	_tmp := strings.Split(csc2.readFchat(), "\n")

	if len(_tmp) != 2 ||
		strings.Compare(_tmp[0], csc1.Username) != 0 ||
		strings.Compare(_tmp[1], mess) != 0 {
		t.Fatalf("chat text message failed.")
	}

	csc2.send2chat(C_chat_select, chat2)
	if strings.Compare(csc2.readFchat(), "Chat doesn't exist") == 0 {
		t.Fatalf("chat %s does not exist.", chat2)
	}

	csc1.send2chat(C_text, mess)

	// reading from csc2 must has timeout
	c1 := make(chan string, 1)
	go func() {
		c1 <- csc2.readFchat()
	}()

	select {
	case <-c1:
		t.Fatalf("message from a different chat.")
		break
	case <-time.After(1 * time.Second):
		break // pass
	}
}
