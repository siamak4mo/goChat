#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
#include <locale.h>
#include <ncurses.h>
#include <unistd.h>
#include <stdbool.h>
#include "chat_window.h"

#define INP_W_LEN 4


static inline int
got_enough_space(struct winsize w)
{
  if (w.ws_row < 10 || w.ws_col < 10)
    return 0;
  else return 1;
}

int
main (void)
{
  struct winsize w;
  ioctl (STDOUT_FILENO, TIOCGWINSZ, &w);
  if (!got_enough_space(w))
    {
      puts ("terminal is too small");
      return -1;
    }

  // init ncurses
  initscr ();
  // make chat window (cw)
  chatw cw = mk_chatw (w.ws_row-INP_W_LEN, w.ws_col, false);
  // initialize cw at (0, 0)
  init_chat_window(&cw, 0, 0);
  
  // make input window
  chatw inpw = mk_chatw (INP_W_LEN, w.ws_col, true);
  inpw.name = "my name";
  
  // initialize cw at (cw.x + 1, 0)
  init_chat_window(&inpw, w.ws_row-INP_W_LEN, 0);

  wchar_t *buf = malloc (500*sizeof(wchar_t));
  memset (buf, 0, 500*sizeof(wchar_t));

  while(!(buf[0]=='E' && buf[1]=='O' && buf[2]=='F'))
    {
      cw_read (&inpw, buf, 500);
      cw_write (&cw, buf);
    }
  endwin ();
  
  return 0;
}
