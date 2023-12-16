this is a zero-dependency Golang project
make it using the existing makefile
you will get server.bin which contains
the chat server itself and the admin server

running the server.bin in the absence of the gochat_server.json
configuration file will lead to the default configuration
of the server which you can find in the gochat_server.template file



  ================
   go chat server
  ================

run and connect to the server using any client you want:

$ ../server.bin
loading configuration failed -- open gochat_server.json: no such file or directory
loading default configuration
Listening on 127.0.0.1:8080
Chat Server | 1702716993 [INFO] chat 4563486fda39a3ee added
Chat Server | 1702716993 [INFO] chat 48434bda39a3ee5e added

connecting to the server using the netcat program:
$ nc localhost 8080
S myname                          # to sign up as myname
Token: bXluYW1l.95a2794a6...      # your token

then you can log in using this token:
L bXluYW1l.95a2794a6...
Logged in

now if you send 'C', the server gives you available chat rooms:
C
ChatID: 4563486fda39a3ee -- Name: EcHo
ChatID: 48434bda39a3ee5e -- Name: HCK
EOF

C 4563486fda39a3ee                # to join the EcHo room

* to send a text message in the joined room:
T hi there

* to change the room:
C 48434bda39a3ee5e                # now you joined the HCK room

* and send 'L' to log out.



  ==============
   admin server
  ==============

it's for monitoring the chat_server itself
by default, it's listening on localhost:8081

$ curl localhost:8081 | jq
{
  "admin server": {
    "Routes": {
       ...
     },
    "name": "Admin Server",
    "address": "127.0.0.1:8081"
  },
  "chat server": {
    "address": "127.0.0.1:8080",
    "status": "OK"
  }
}

there is a brief documentation of the API in the `Routes` section

* to see the current chat_server configuration:
$ curl 127.0.0.1:8081/config/lookup | jq
{
  "Token": {
    ...
  },
  "Admin": {
    "admin_addr": "127.0.0.1:8081"
  },
  "Server": {
    "listen_addr": "127.0.0.1:8080",
    "room_names": ["EcHo", "HCK"],
    "room_motds": ["Welcome to the `echo` chat!", "Welcome to the `Hack` chat :D"]
  },
  "Log": {"log_level": 0}
}
now server.bin file will log a warning:
Admin Server| 1702717840 [Warning] Config Lookup Access

* to add a new chat room:
$ curl 127.0.0.1:8081/chat/add -XPOST -d '{"name": "chat name", "banner": "MOTD"}'

* to remove a chat room:
$ curl 127.0.0.1:4242/chat/remove -XPOST -d '{"chat key": "36393639da39a3ee"}'

* to see users of a chat room and statistics:
$ curl "127.0.0.1:8081/chat/users?chat=4563486fda39a3ee"
{"0":"user1", "1":"user2", ...}

$ curl 127.0.0.1:8081/chats/stat | jq    
{
  "4563486fda39a3ee": {
    "name": "EcHo",
    "banner": "Welcome to the `echo` chat!",
    "member count": 2
  },
   ...
}
