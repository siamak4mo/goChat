#GO_CC = -compiler gccgo  # uncomment to use gccgo
GO_BUILD = go build $(GO_CC)
BUILD_OUT = ../server.bin

all: $(BUILD_OUT)


$(BUILD_OUT): go.mod server/server.go server/token.go server_main.go
	$(GO_BUILD) -o $(BUILD_OUT) server_main.go
