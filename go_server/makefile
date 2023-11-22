#GO_CC = -compiler gccgo  # uncomment to use gccgo
GO = go $(GO_CC)
BUILD = ../server.bin

all: go.mod server/server.go server/token.go server_main.go
	go build -o $(BUILD) server_main.go
