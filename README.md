# Go Chat (gochat)

go_server: server program  --  c_client: client program

## Make the project

run `make deploy` to build `server.bin`, `client.bin`, and
`gochat_server.json` (the server configuration) files.

the server program is a zero-dependency `GoLang` project,
and the client program is written in `C` and only depends
on the `ncurses` library, so be sure `ncurses.h` is available.

edit `gochat_server.json` to change the default configuration
of the server.

---

## running and testing the server

run `./server.bin`, it will try to find the `gochat_server.json`
file in your current working directory,
otherwise run `./server.bin -C /path/to/config.json`.

run `make test_server` to run tests for the server

## using the client

run `./client.bin -s "SERVER_ADDR" -p "SERVER_PORT" -u my_name`
if you see `* login token` it means you logged in successfully
and the file `/tmp/client.config` will be generated,
so the next time, you can log in using this file: 
`./client.bin -c /tmp/client.config`
