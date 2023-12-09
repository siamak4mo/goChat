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
static int rxoff, ryoff; // inpw (x,y) cursor offset
static chat_net cn;
static bool isJoined = false;
static char chatID[17];

struct Options {
  char server_addr[16];
  int server_port;
  char *user_token;
  char *username;
  bool EOO; // end of options
};
static struct Options opt = {
  .EOO = false,
};

typedef enum {
  Uninitialized = 1,
  Initialized,
  Signedup,
  Logedin,
  Joined
} Cstate;
static Cstate state = Uninitialized;

#define ERR_LOAD_CONFIG -1
#define ERR_UNKNOWN_ARG 1

#define ST_CURSOR() getyx (inpw.w, ryoff, rxoff);
#define LD_CURSOR() wmove (inpw.w, ryoff, rxoff);
#define SAFE_CALL(fun_call) do {                \
    ST_CURSOR();                                \
    fun_call;                                   \
    LD_CURSOR();                                \
    wrefresh (inpw.w);         } while (0)

// config file handling
static const char default_config_path[] = "/tmp/client.config";
static inline int load_config_from_file(const char *path);
static inline int save_config(const char *path);

static void
__exit()
{
  // free ncurses mem
  endwin ();
  // free chat window memory
  cw_end (&cw);
  cw_end (&inpw);
  // free opt memory
  if (opt.username != NULL)
    free (opt.username);
  if (opt.user_token != NULL)
    free (opt.user_token);
  // free chat_net mem
  net_end (&cn);
}

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

  // init ncurses
  initscr ();
  // initialize cw at (0, 0)
  init_chat_window(&cw, 0, 0);
  // initialize cw at (cw.x + 1, 0)
  init_chat_window(&inpw, w.ws_row-INP_W_LEN, 0);
  // end of GUI initialization
  state = Initialized;

  wchar_t *buf = malloc (MAX_BUF*sizeof(wchar_t));
  memset (buf, 0, MAX_BUF*sizeof(wchar_t));

  while (1)
    {
      cw_read (&inpw, buf, MAX_BUF);
      if (buf[0]=='\0')
        continue;
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
              wcharcpy(chatID, buf);
              cw_set_name (&cw, chatID);
            }
        }
      else if (state == Joined)
        { // send text packet
          net_wwrite (&cn, TEXT, buf);
          
          SAFE_CALL(cw_write_my_mess(&cw, buf));
        }
    }
  free (buf);
  return 0;
}

static inline int
NETWORK_loop_H(void *)
{
  int n;
  char *p;
  cn = net_new ();

  while (state != Initialized) {};
  // init connection to the server
  if (strlen (opt.server_addr) == 0 ||
    net_init (&cn, opt.server_addr, opt.server_port) != 0)
    {
      SAFE_CALL(cw_write_char (&cw, " ? could not connect to the server - exiting"));
      return -1;
    }
  SAFE_CALL(cw_write_char (&cw, " * connected to the server"));
  state = Initialized;

  if (opt.user_token == NULL || strlen (opt.user_token) == 0)
    { // to signup
      if (opt.username == NULL || strlen (opt.username) == 0)
        {
          SAFE_CALL(cw_write_char (&cw, " ? either a token or username is expected - exiting"));
          return -1;
        }
      net_write (&cn, SIGNUP, opt.username, strlen (opt.username));
      p = net_read (&cn, &n);
      if (strncmp(p, "Token: ", 7) != 0)
        {
          SAFE_CALL(cw_write_char (&cw, " ? failed to signup - exiting"));
          return -1;
        }
      else
        {
          opt.user_token = malloc (n-7);
          strncpy (opt.user_token, p+7, n-7);
          state = Signedup;
        }
    }

  // begin to login
  net_write (&cn, LOGIN_OUT, opt.user_token, strlen (opt.user_token));
  p = net_read (&cn, &n);
  if (strncmp(p, "Loged in", 8) != 0)
    {
      SAFE_CALL(cw_write_char (&cw, " ? failed to login - exiting"));
      return -1;
    }
  else
    {
      SAFE_CALL(cw_vawrite_char (&cw, 2, " * login token: ", opt.user_token));
      state = Logedin;
      cw_set_name (&inpw, opt.username);
    }

  // save configuration to the default path
  save_config (default_config_path);

  // begin to select chat to join
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

static inline int
load_config_from_file(const char *path)
{
  int res = 0;
  FILE *f;

  if (path == NULL || strlen (path) == 0)
      f = fopen (default_config_path, "r");
  else
    f = fopen (path, "r");
  
  if (f==NULL)
    return ERR_LOAD_CONFIG;
  
  char *key = malloc (32);
  char *val = malloc (128);
  
  while (fscanf(f, "%32[^ ] %128[^\n]%*c", key, val) == 2)
    {
      if (!strcmp (key, "server_addr"))
        strcpy (opt.server_addr, val);
      else if (!strcmp (key, "server_port"))
        opt.server_port = atoi (val);
      else if (!strcmp (key, "username"))
        {
          opt.username = malloc (strlen (val));
          strcpy (opt.username, val);
        }
      else if (!strcmp (key, "user_token"))
        {
          opt.user_token = malloc (strlen (val));
          strcpy (opt.user_token, val);
        }
      else
        {
          res = ERR_LOAD_CONFIG;
          break;
        }
    }
  
  free (key);
  free (val);
  fclose (f);
  return res;
}

static inline int
save_config(const char *path)
{
  FILE *f = fopen (path, "w+");
  if (f==NULL)
    return 1;
  fprintf (f, "server_addr %s\n", opt.server_addr);
  fprintf (f, "server_port %d\n", opt.server_port);
  fprintf (f, "username %s\n", opt.username);
  fprintf (f, "user_token %s\n", opt.user_token);
  fclose (f);
  return 0;
}

static inline int
get_arg(const char *flag, char *arg)
{
  if (!strcmp (flag, "--"))
    opt.EOO = true;
  else if (!strcmp (flag, "-s") || !strcmp (flag, "--server"))
    strcpy (opt.server_addr, arg);
  else if (!strcmp (flag, "-p") || !strcmp (flag, "--port"))
    opt.server_port = atoi (arg);
  else if (!strcmp (flag, "-u") || !strcmp (flag, "--username"))
    {
      opt.username = malloc (strlen (arg));
      strcpy (opt.username, arg);
    }
  else if (!strcmp (flag, "-t") || !strcmp (flag, "--token"))
    {
      opt.user_token = malloc (strlen (arg));
      strcpy (opt.user_token, arg);
    }
  else if (!strcmp (flag, "-c") || !strcmp (flag, "--config"))
    return load_config_from_file (arg);
  else
    return ERR_UNKNOWN_ARG;
  return 0;
}

static inline int
pars_args(int argc, char **argv)
{
  for (argc--, argv++; argc > 0; argc--, argv++)
    {
      if (!opt.EOO && argv[0][0] == '-')
        {
          if (get_arg (*argv, *(argv+1)) == ERR_UNKNOWN_ARG)
            {
              printf ("unknown argument %s\n", argv[0]);
              return 1;
            }
          if (get_arg (*argv, *(argv+1)) == ERR_LOAD_CONFIG)
            {
              printf ("loading config from file failed\n");
              return -1;
            }
        }
      else
        {
          // dash-dash prefixed
        }
    }
  return 0;
}

int
main(int argc, char **argv)
{
  if (pars_args (argc, argv) != 0)
    return 1;
  if (!got_enough_space(w))
    {
      puts ("terminal is too small");
      return 1;
    }
  
  thrd_t t;
  thrd_create (&t, NETWORK_loop_H, NULL);
  GUI_loop_H (NULL);

  thrd_detach (t);
  __exit ();
  
  return 0;
}
