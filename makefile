SERVER_SRC = go_server
CLIENT_SRC = c_client

all: build_server build_client

deploy: build_server build_client
	cp $(SERVER_SRC)/gochat_server.template gochat_server.json

## server ##
.PHONY: build_server
build_server:
	$(MAKE) -C $(SERVER_SRC)

.PHONY: run_server
run_server:
	$(MAKE) -C $(SERVER_SRC) run

.PHONY: test_server
test_server:
	$(MAKE) -C $(SERVER_SRC) test

## client ##
.PHONY: build_client
build_client:
	$(MAKE) -C $(CLIENT_SRC)
