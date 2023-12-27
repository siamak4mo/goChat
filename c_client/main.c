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
static char *ERR_MSG;

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
  Joined
} Cstate;
static Cstate state = Uninitialized;
static bool cw_lock = false;

#define ST_CURSOR() getyx (inpw.w, ryoff, rxoff);
#define LD_CURSOR() wmove (inpw.w, ryoff, rxoff);
#define SAFE_CW_WRITE(fun_call) do {            \
    while (cw_lock) {};                         \
    cw_lock = true;                             \
    ST_CURSOR();                                \
    fun_call;                                   \
    LD_CURSOR();                                \
    cw_lock = false;                            \
    wrefresh (inpw.w);             } while (0)

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
  else
    {
      ERR_MSG = "terminal is too smal -- exiting.";
      return 1;
    }
}

static inline void
select_chat()
{
  int n;
  char *p;
  // begin to select chat to join
  net_write (&cn, CHAT_SELECT, NULL, 0);
  SAFE_CW_WRITE(cw_write_char (&cw, " * type chatID to join..."));
  
  while (cn.state == Connected)
    {
      p = net_read (&cn, &n);
      if (strncmp (p, "EOF", 3) == 0)
        break;
      else if (n>4 && strncmp (p+n-4, "EOF", 3) == 0) 
        {
          p[n-4] = '\0';
          SAFE_CW_WRITE(cw_write_char (&cw, p));
          break;
        }
      else
        SAFE_CW_WRITE(cw_write_char (&cw, p));
    }
}

static inline int
connect_to_server()
{
 if (strlen (opt.server_addr) == 0 ||
    net_init (&cn, opt.server_addr, opt.server_port) != 0)
    {
      ERR_MSG = "Could not connect to the server - exiting";
      return -1;
    }
  SAFE_CW_WRITE(cw_write_char (&cw, " * connected to the server"));
  state = Initialized;

  return 0;
}

static inline int
try_to_login()
{
  int n;
  char *p;
  
  if (opt.user_token == NULL || strlen (opt.user_token) == 0)
    { // to signup
      if (opt.username == NULL || strlen (opt.username) == 0)
        {
          ERR_MSG = "Either a token or username is required - exiting";
          return -1;
        }
      net_write (&cn, SIGNUP, opt.username, strlen (opt.username));
      p = net_read (&cn, &n);
      if (strncmp (p, "Token: ", 7) != 0)
        {
          if (strncmp (p, "User Already exists", 19) == 0)
            ERR_MSG = "failed to signup, user already exists -- exiting";
          else
            ERR_MSG = "failed to signup -- exiting";
          return -1;
        }
      else
        {
          opt.user_token = malloc (n-7);
          strncpy (opt.user_token, p+7, n-7);
        }
    }

  // begin to login
  net_write (&cn, LOGIN_OUT, opt.user_token, strlen (opt.user_token));
  p = net_read (&cn, &n);
  if (strncmp(p, "Logged in", 8) != 0)
    {
      ERR_MSG = "failed to login - exiting";
      return -1;
    }
  else
    {
      SAFE_CW_WRITE(cw_vawrite_char (&cw, 2, " * login token: ", opt.user_token));
      if (opt.username == NULL)
        {
          // get the username from the server
          net_write (&cn, WHOAMI, NULL, 0);
          p = net_read (&cn, &n);
          if (n>11 && strncmp (p, "Username: ", 10) == 0)
            {
              int uname_len = 0;
              while (p[uname_len] != '\0' && p[uname_len] != '\n')
                uname_len++;
              p[uname_len] = '\0';
              opt.username = malloc(uname_len);
              strncpy (opt.username, p+10, uname_len);
            }
          else
            cw_set_name (&inpw, "Unknown Error");
        }
      cw_set_name (&inpw, opt.username);
    }
  return 0;
}

static inline int
MAIN_loop_H (void *)
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
  cn = net_new ();
  // init connection to the server
  if (connect_to_server () != 0)
    return -1;
  // try to login
  if (try_to_login () != 0)
    return -1;
  // save configuration to the default path
  save_config (default_config_path);
  // print chat rooms and ask user to select one
  select_chat ();
  
  wchar_t *buf = malloc (MAX_BUF*sizeof(wchar_t));
  memset (buf, 0, MAX_BUF*sizeof(wchar_t));
  
  while (cn.state == Connected)
    {
      cw_read (&inpw, buf, MAX_BUF);
      if (buf[0]=='\0')
        continue;
      if (buf[0]=='E' && buf[1]=='O' && buf[2]=='F')
        break;
      
      if (state != Joined)
        { // send chat select packet
          net_wwrite (&cn, CHAT_SELECT, buf);
          const char *p = net_read (&cn, NULL);
          if (strncmp (p, "Chat doesn't exist", 18) == 0)
            SAFE_CW_WRITE(cw_write_char (&cw, " ? chat not found - try again"));
          else if (strlen (p) != 0)
            {
              isJoined = true;
              cw_clear (&cw);
              SAFE_CW_WRITE(cw_vawrite_char (&cw, 2, p, "  --  (*) is you"));
              state = Joined;
              wcharcpy(chatID, buf);
              cw_set_name (&cw, chatID);
            }
        }
      else
        { // send text packet
          net_wwrite (&cn, TEXT, buf);
          SAFE_CW_WRITE(cw_write_my_mess(&cw, buf));
        }
    }
  free (buf);
  return 0;
}

static inline int
READCHAT_loop_H(void *)
{
  int n;
  char *p;
  
  while (state != Joined) {};
  while(cn.state == Connected)
    { // text message
      p = net_read (&cn, &n);
      if (strlen (p) != 0)
        SAFE_CW_WRITE(cw_write_char_mess (&cw, p));
    }

  return 0; // unreachable
}

static inline int
load_config_from_file(const char *path)
{
  int res = 0, r = 2;
  FILE *f;

  if (path == NULL || strlen (path) == 0)
      f = fopen (default_config_path, "r");
  else
    f = fopen (path, "r");
  
  if (f==NULL)
    {
      ERR_MSG = "Could not open config file -- exiting.";
      return -1;
    }
  
  char *key = malloc (32);
  char *val = malloc (128);
  
  while (r == 2)
    {
      r = fscanf(f, "%32[^ ] %128[^\n]%*c", key, val);
      if (r < 0)
        {
          ERR_MSG = "Could not read the config file -- exiting.";
          return -1;
        }
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
          ERR_MSG = "parsing config file failed -- exiting.";
          return -1;
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
    {
      if (!load_config_from_file (arg))
        return 1;
    }
  else
    {
      ERR_MSG = "unknown argument -- exiting.";
      return 1;
    }
  return 0;
}

static inline int
pars_args(int argc, char **argv)
{
  for (argc--, argv++; argc > 0; argc--, argv++)
    {
      if (!opt.EOO && argv[0][0] == '-')
        {
          if (get_arg (*argv, *(argv+1)) != 0)
            return -1;
          if (get_arg (*argv, *(argv+1)) != 0)
            return -1;
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
    {
      puts (ERR_MSG);
      return 1;
    }
  if (!got_enough_space(w))
    {
      puts (ERR_MSG);
      return 1;
    }
  
  thrd_t t;
  if (thrd_create (&t, READCHAT_loop_H, NULL) < 0)
    {
      puts ("Could not allocate thread -- exiting.");
      return 1;
    }
  MAIN_loop_H (NULL);

  thrd_detach (t);
  __exit ();

  if (ERR_MSG != NULL)
    puts (ERR_MSG);
  return 0;
}
