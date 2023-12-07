CFLAGS=-g -ggdb -Wall -Wextra `pkg-config --cflags ncurses` -I.
LDFLAGS=`pkg-config --libs ncurses`
BUILD_OUT=../client.bin
OBJ=net.o chat_window.o


all: main.c $(OBJ)
	$(CC) $(CFLAGS) -c main.c -o main.o
	$(CC) $(CFLAGS) -o $(BUILD_OUT) main.o $(OBJ) $(LDFLAGS)

%.o: %.c %.h
	$(CC) $(CFLAGS) -c $< -o $@


.PHONY: clean
clean:
	rm $(OBJ) $(BUILD_OUT)
