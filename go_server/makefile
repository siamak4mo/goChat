#GO_CC = -compiler gccgo  # uncomment to use gccgo
GO_BUILD = go build $(GO_CC)
GO_RUN = go run $(GO_CC)
BUILD_OUT = ../server.bin

SRC = go.mod server/server.go server/stoken/token.go server/config/config.go server_main.go

all: $(BUILD_OUT)


$(BUILD_OUT): $(SRC) 
	$(GO_BUILD) -o $(BUILD_OUT) server_main.go


.PHONY: run
run:
	$(GO_RUN) server_main.go
