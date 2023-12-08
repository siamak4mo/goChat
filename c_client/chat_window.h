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

static const wchar_t ME[] = {'(', '*', ')', '|', ' ', '\0'};
#define cw_write_my_mess(cw, buf) cw_vawrite (cw, 2, ME, buf);

void init_chat_window(chatw *, int, int);
void cw_clear(chatw *);
void cw_write(chatw *, const wchar_t *);
void cw_write_char(chatw *, const char *);
void cw_write_char_mess(chatw *, const char *);
void cw_vawrite(chatw *, int, ...);
void cw_vawrite_char(chatw *, int, ...);
chatw mk_chatw(int, int, bool);
int cw_read(chatw *, wchar_t *, int);
#endif
