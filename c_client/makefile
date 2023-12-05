CFLAGS=-Wall -Wextra `pkg-config --cflags ncurses` -I.
LDFLAGS=`pkg-config --libs ncurses`
BUILD_OUT=../client.bin
OBJ=main.o chat_window.o


all: $(OBJ)
	$(CC) -o $(BUILD_OUT) $(OBJ) $(LDFLAGS)

%.o: %.c %.h
	$(CC) $(CFLAGS) -c $< -o $@


.PHONY: clean
clean:
	rm $(OBJ) $(BUILD_OUT)
