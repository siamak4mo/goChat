# Go Chat (gochat)

go_server: server program  --  c_client: client program

## Make the project

run `make deploy` to build `server.bin`, `client.bin`, and
`gochat_server.json` (the server configuration) files.

the provided makefile uses `gccgo` compiler, if you want to use
the default go compiler, comment the first line of the `go_server/makefile` out.

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

### make token for users
users only can login with not trusted usernames (NT_xxx names)
unless you make trusted tokens from the admin server
see `go_server/README.txt` 'admin server' section for more information

## using the client

run `./client.bin -s "SERVER_ADDR" -p "SERVER_PORT" -u NT_my_name`
if you see `* login token` it means you logged in successfully
and the file `/tmp/client.config` will be generated,
so the next time, you can log in using this file: 
`./client.bin -c /tmp/client.config`

if you have a trusted token from admin, you can login by:
`./client.bin -s "SERVER_ADDR" -p "SERVER_PORT" -t "YOUR_TOKEN"`
