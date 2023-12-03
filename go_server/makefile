#GO_CC = -compiler gccgo  # uncomment to use gccgo
GO_BUILD = go build $(GO_CC)
GO_RUN = go run $(GO_CC)
GO_TEST = go test

BUILD_OUT = ../server.bin
TEST_DIR  = test

SRC = go.mod server/server.go server/stoken/token.go server/config/config.go server_admin.go server_main.go

all: $(BUILD_OUT)


$(BUILD_OUT): $(SRC) 
	$(GO_BUILD) -o $(BUILD_OUT) server_admin.go server_main.go


.PHONY: run
run:
	$(GO_RUN) server_admin.go server_main.go


.PHONY: test
test:
	cd $(TEST_DIR); $(GO_TEST)
