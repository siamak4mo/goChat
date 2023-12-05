all: main.c
	$(CC) -c chat_window.c -I.
	$(CC) -c main.c -I.
	$(CC) -o ../client.bin main.o chat_window.o -lncursesw
