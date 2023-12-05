#ifndef CHAT_W__
#define CHAT_W__
#include <stdbool.h>
#include <stddef.h>

typedef struct {
  int row, col;
  int padding;
  bool box;
  int line_c;
  char *name;
  WINDOW *w;
} chatw;


void init_chat_window(chatw *, int, int);
void cw_write(chatw *, const wchar_t *);
chatw mk_chatw(int, int, bool);
int cw_read(chatw *, wchar_t *, int);
#endif
