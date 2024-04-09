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

#define ST_CUR(win, ry, rx) getyx (win, ry, rx);
#define LD_CUR(win, ry, rx) wmove (win, ry, rx);
/* use this macro to run procedures that might involve */
/* moving the cursor, pass __DO__ like {... f(x); ...} */
#define SAFE_CW(win, __DO__) do {                       \
    int __y, __x;                                       \
    ST_CUR(win, __y, __x);                              \
    __DO__;                                             \
    LD_CUR(win, __y, __x);                              \
    wrefresh(win); } while (0)

static const wchar_t __ME[] = {'(', '*', ')', '|', ' ', '\0'};
/* use this macro to show messages written by the user */
#define cw_write_my_mess(cw, mess_buf)                  \
  cw_vawrite (cw, 2, __ME, mess_buf);


/* create and initialize chat window */
chatw mk_chatw(int, int, bool);
void init_chat_window(chatw *, int, int);
/* set window name (displayed on the top-right corner) */
void cw_set_name(chatw *cw, const char *name);
/* clear contents */
void cw_clear(chatw *);
/* write text to chat window functions */
void cw_write(chatw *, const wchar_t *);
void cw_write_char(chatw *, const char *);
void cw_write_char_mess(chatw *, const char *);
void cw_vawrite(chatw *, int, ...);
void cw_vawrite_char(chatw *, int, ...);
/* read from chat window functions */
int cw_read(chatw *, wchar_t *, int);
void cw_end (chatw *);
#endif
