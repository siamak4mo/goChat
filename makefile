SERVER_SRC = go_server

all: build_server

deploy: build_server
	cp $(SERVER_SRC)/gochat_server.template gochat_server.json

.PHONY: build_server
build_server:
	$(MAKE) -C $(SERVER_SRC)


.PHONY: run_server
run_server:
	$(MAKE) -C $(SERVER_SRC) run


.PHONY: test_server
test_server:
	$(MAKE) -C $(SERVER_SRC) test

