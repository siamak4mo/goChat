package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	LADDR = "127.0.0.1"
	LPORT = ":8080"
)



func root_handler(w http.ResponseWriter, r *http.Request){
	if auth, ok := r.Header["Authorization"]; ok {
		tk, err := Init_btoken(auth[0])

		if err != nil{
			w.Write([]byte(err.Error()))
		}else{
			if tk.Validate(){
				w.Write([]byte(tk.Username))
			}else{
				w.Write([]byte("Invalid Token\n"))
			}
		}
	}else{
		w.Write([]byte("go to login\n"))
	}
}

func login_handler(w http.ResponseWriter, r *http.Request){
	tk := Init_token()
	tk.Username = []byte(r.URL.Query().Get("username"))
	tk.MkToken()

	w.Write([]byte(tk.Token))
}


func main(){
	fmt.Println("RUNNING -- server.go")
	
	http.HandleFunc("/login", login_handler)
	http.HandleFunc("/", root_handler)

	
	log.Printf("Listening on %s%s", LADDR, LPORT)
	err := http.ListenAndServe(LADDR + LPORT, nil)
	if err != nil{
		log.Fatal(err)
	}
}
