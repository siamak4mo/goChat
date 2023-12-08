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
#include "net.h"

#define INP_W_LEN 4
#define MIN_W_LEN 10
#define MAX_BUF 500

static struct winsize w;
static chatw cw, inpw;
static bool GUI_II = false; // gui is initialized
static int rxoff, ryoff; // inpw (x,y) cursor offset
static chat_net cn;
static char server_addr[] = "127.0.0.1";
static int server_port = 7070;
static char *user_token;
static char username[] = "my name";
static bool isJoined = false;
static char chatID[17];

typedef enum {
  Uninitialized = 1,
  Initialized,
  Signedup,
  Logedin,
  Joined
} Cstate;
static Cstate state = Uninitialized;

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

  while (1)
    {
      cw_read (&inpw, buf, MAX_BUF);
      if (buf[0]=='E' && buf[1]=='O' && buf[2]=='F')
        break;
      
      if (state == Logedin)
        { // send chat select packet
          net_wwrite (&cn, CHAT_SELECT, buf);
          const char *p = net_read (&cn, NULL);
          if (strncmp (p, "Chat doesn't exist", 18) == 0)
            SAFE_CALL(cw_write_char (&cw, " ? chat not found - try again"));
          else
            {
              isJoined = true;
              cw_clear (&cw);
              SAFE_CALL(cw_vawrite_char (&cw, 2, p, "  --  (*) is you"));
              state = Joined;
            }
        }
      else if (state == Joined)
        { // send text packet
          net_wwrite (&cn, TEXT, buf);
          
          SAFE_CALL(cw_write_my_mess(&cw, buf));
        }
    }
  endwin ();
  free (buf);
  return 0;
}

static inline int
NETWORK_loop_H(void *)
{
  int n;
  char *p;
  cn = net_new ();
  int res = net_init (&cn, server_addr, server_port);

  if (res != 0)
    return res;
  SAFE_CALL(cw_write_char (&cw, " * connected to the server"));
  state = Initialized;

  net_write (&cn, SIGNUP, username, strlen (username));
  p = net_read (&cn, &n);
  if (strncmp(p, "Token: ", 7) != 0)
    {
      SAFE_CALL(cw_write_char (&cw, " ? failed to signup - exiting"));
      return -1;
    }
  else
    {
      user_token = malloc (n-7);
      strncpy (user_token, p+7, n-7);
      state = Signedup;
    }
  
  net_write (&cn, LOGIN_OUT, user_token, strlen (user_token));
  p = net_read (&cn, &n);
  if (strncmp(p, "Loged in", 8) != 0)
    {
      SAFE_CALL(cw_write_char (&cw, " ? failed to login - exiting"));
      return -1;
    }
  else
    {
      SAFE_CALL(cw_write_char (&cw, " * loged in"));
      state = Logedin;
    }

  net_write (&cn, CHAT_SELECT, NULL, 0);
  SAFE_CALL(cw_write_char (&cw, " * type chatID to join..."));
  
  while (1)
    {
      p = net_read (&cn, &n);
      if (strncmp (p, "EOF", 3) == 0)
        break;
      else if (n>4 && strncmp (p+n-4, "EOF", 3) == 0) 
        {
          p[n-4] = '\0';
          SAFE_CALL(cw_write_char (&cw, p));
          break;
        }
      else
        SAFE_CALL(cw_write_char (&cw, p));
    }

  while (state != Joined) {};
  while(1)
    {
      // text message
      p = net_read (&cn, &n);
      SAFE_CALL(cw_write_char_mess (&cw, p));
    }

  net_end (&cn);
  return 0; // unreachable
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
  NETWORK_loop_H(NULL);
  
  return 0;
}
