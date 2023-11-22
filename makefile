SERVER_SRC = go_server

all: build_server


.PHONY: build_server
build_server:
	$(MAKE) -C $(SERVER_SRC)
