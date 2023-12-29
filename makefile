SERVER_SRC = go_server
CLIENT_SRC = c_client

SERVER_BIN = server.bin
CLIENT_BIN = client.bin

all: build_server build_client

deploy: build_server build_client
	cp $(SERVER_SRC)/gochat_server.template gochat_server.json

## server ##
.PHONY: build_server
build_server: | $(SERVER_BIN)
	ln -s $(SERVER_SRC)/$(SERVER_BIN) .
$(SERVER_BIN):
	$(MAKE) -C $(SERVER_SRC)

.PHONY: run_server
run_server:
	$(MAKE) -C $(SERVER_SRC) run

.PHONY: test_server
test_server:
	$(MAKE) -C $(SERVER_SRC) test

## client ##
.PHONY: build_client
build_client: | $(CLIENT_BIN)
	ln -s $(CLIENT_SRC)/$(CLIENT_BIN) .
$(CLIENT_BIN):
	$(MAKE) -C $(CLIENT_SRC)

.PHONY: clean
clean:
	$(MAKE) -C $(CLIENT_SRC) clean
	rm server.bin
	rm gochat_server.json
