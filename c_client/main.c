#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
#include <locale.h>
#include <ncurses.h>
#include <unistd.h>
#include <stdbool.h>
#include <threads.h>
#include "chat_window.h"

#define INP_W_LEN 4
#define MIN_W_LEN 10
#define MAX_BUF 500

static struct winsize w;
static chatw cw, inpw;
static bool GUI_II = false; // gui is initialized
static int rxoff, ryoff; // inpw (x,y) cursor offset

#define ST_CURSOR() getyx (inpw.w, ryoff, rxoff);
#define LD_CURSOR() wmove (inpw.w, ryoff, rxoff);
#define SAFE_CALL(fun_call) do {                \
    ST_CURSOR();                                \
    fun_call;                                   \
    LD_CURSOR();                                \
    wrefresh (inpw.w);         } while (0)

static inline int
got_enough_space()
{
  ioctl (STDOUT_FILENO, TIOCGWINSZ, &w);

  if (w.ws_row < MIN_W_LEN || w.ws_col < MIN_W_LEN)
    return 0;
  else return 1;
}

static inline int
GUI_loop_H (void *)
{
  // make chat and input windows
  cw = mk_chatw (w.ws_row-INP_W_LEN, w.ws_col, false);
  inpw = mk_chatw (INP_W_LEN, w.ws_col, true);
  inpw.name = "my name";

  // init ncurses
  initscr ();
  // initialize cw at (0, 0)
  init_chat_window(&cw, 0, 0);
  // initialize cw at (cw.x + 1, 0)
  init_chat_window(&inpw, w.ws_row-INP_W_LEN, 0);
  // end of GUI initialization
  GUI_II = true;

  wchar_t *buf = malloc (MAX_BUF*sizeof(wchar_t));
  memset (buf, 0, MAX_BUF*sizeof(wchar_t));

  while(!(buf[0]=='E' && buf[1]=='O' && buf[2]=='F'))
    {
      cw_read (&inpw, buf, MAX_BUF);
      cw_write (&cw, buf);
    }
  endwin ();
  return 0;
}

int
main(void)
{
  if (!got_enough_space(w))
    {
      puts ("terminal is too small");
      return -1;
    }
  
  thrd_t t;
  thrd_create (&t, GUI_loop_H, NULL);

  while(!GUI_II){};
  // do other stuff
  SAFE_CALL(cw_write_char (&cw, "GUI was initialized."));

  while(1){};

  return 0;
}
