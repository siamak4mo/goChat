_CFLAGS_H=-g -ggdb
# _CFLAGS_H=-O3 # uncomment to enable optimization

CFLAGS=$(_CFLAGS_H) -Wall -Wextra -Werror `pkg-config --cflags ncurses` -I.
LDFLAGS=`pkg-config --libs ncurses`
BUILD_OUT=../client.bin
OBJ=net.o chat_window.o


all: $(BUILD_OUT)

$(BUILD_OUT): main.o $(OBJ)
	$(CC) $(CFLAGS) -o $(BUILD_OUT) main.o $(OBJ) $(LDFLAGS)

main.o: main.c
	$(CC) $(CFLAGS) -c $< -o $@

%.o: %.c %.h
	$(CC) $(CFLAGS) -c $< -o $@


.PHONY: clean
clean:
	rm main.o $(OBJ) $(BUILD_OUT)
