_CFLAGS_H=-g -ggdb
# _CFLAGS_H=-O3 # uncomment to enable optimization

CFLAGS=$(_CFLAGS_H) -Wall -Wextra -Werror `pkg-config --cflags ncurses` -I.
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
	rm main.o $(OBJ) $(BUILD_OUT)
