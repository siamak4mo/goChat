#GO_CC = -compiler gccgo  # uncomment to use gccgo
GO_BUILD = go build $(GO_CC)
GO_RUN = go run $(GO_CC)
GO_TEST = go test

BUILD_OUT = ../server.bin
TEST_DIR  = test
SRC_DIR   = chat_server

SRC = go.mod $(SRC_DIR)/server.go $(SRC_DIR)/stoken/token.go config/config.go logger/logger.go admin_server.go server_main.go

all: $(BUILD_OUT)


$(BUILD_OUT): $(SRC) 
	$(GO_BUILD) -o $(BUILD_OUT) admin_server.go server_main.go


.PHONY: run
run:
	$(GO_RUN) admin_server.go server_main.go


.PHONY: test
test:
	cd $(TEST_DIR); $(GO_TEST)
