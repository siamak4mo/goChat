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
static chat_net cn;
static bool isJoined = false;
static char chatID[17];
static char *ERR_MSG;
static char *__progname__;

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
static bool window_lock = false;

#define RET_ERR(msg, code) do { ERR_MSG = (msg); return code; } while (0)

#define SAFE_WRITE(__DO__) do {                     \
    while (window_lock) {};                         \
    window_lock = true;                             \
    SAFE_CW (inpw.w, {                              \
        __DO__;                                     \
      });                                           \
    window_lock = false;                            \
    wrefresh (inpw.w); } while (0)

// config file handling
static const char default_config_path[] = "/tmp/client.config";
static inline int load_config_from_file (const char *path);
static inline int save_config (const char *path);

static void
__exit ()
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

static inline void
Usage ()
{
  printf ("Usage: %s [OPTIONS]\n%s", __progname__,
          "OPTIONS:\n"
          "   -s, --server         to specify server address\n"
          "   -p, --port           to specify server listening port\n"
          "   -u, --username       to specify username to sign up (as not trusted user)\n"
          "   -t, --token          to use your having token\n"
          "   -c, --config         to specify config file path\n"
  );
}

static inline int
got_enough_space ()
{
  ioctl (STDOUT_FILENO, TIOCGWINSZ, &w);

  if (w.ws_row < MIN_W_LEN || w.ws_col < MIN_W_LEN)
    RET_ERR ("Terminal is too small -- Exiting.", -1);
  else
    return 0;
}

static inline void
select_chat__H ()
{
  int n;
  char *p;
  // begin to select chat to join
  net_write (&cn, CHAT_SELECT, NULL, 0);
  SAFE_WRITE (cw_write_char (&cw, " * enter chatID to join..."));
  
  while (cn.state == Connected)
    {
      p = net_read (&cn, &n);
      if (strncmp (p, "EOF", 3) == 0)
        break;
      else if (n>4 && strncmp (p+n-4, "EOF", 3) == 0) 
        {
          p[n-4] = '\0';
          SAFE_WRITE (cw_write_char (&cw, p));
          break;
        }
      else
        SAFE_WRITE (cw_write_char (&cw, p));
    }
}

static inline int
conn_to_server__H ()
{
  if (strlen (opt.server_addr) == 0 ||
      net_init (&cn, opt.server_addr, opt.server_port) != 0)
    RET_ERR ("Could Not Connect to the Server -- Exiting", -1);
  SAFE_WRITE (cw_write_char (&cw, " * connected to the server"));
  state = Initialized;

  return 0;
}

static inline int
login__H ()
{
  int n;
  char *p;
  
  if (opt.user_token == NULL || strlen (opt.user_token) == 0)
    { // to signup
      if (opt.username == NULL || strlen (opt.username) == 0)
        RET_ERR ("Either a Token or Username is Required -- Exiting", -1);
      net_write (&cn, SIGNUP, opt.username, strlen (opt.username));
      p = net_read (&cn, &n);
      if (strncmp (p, "Token: ", 7) != 0)
        {
          if (strncmp (p, "User Already exists", 19) == 0)
            RET_ERR ("Failed to Signup, User Already Exists -- Exiting", -1);
          else
            RET_ERR ("Failed to Signup -- Exiting", -1);
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
  if (strncmp (p, "Logged in", 8) != 0)
    RET_ERR ("Failed to Login, Token is Not Valid -- Exiting", -1);
  else
    {
      SAFE_WRITE (cw_vawrite_char (&cw, 2, " * token: ", opt.user_token));
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
              opt.username = malloc (uname_len);
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
  cw = mk_chatw (w.ws_row - INP_W_LEN - 1, w.ws_col, false);
  inpw = mk_chatw (INP_W_LEN, w.ws_col, true);
  // init ncurses
  initscr ();
  // initialize cw at (0, 0)
  init_chat_window (&cw, 0, 0);
  // initialize cw at (cw.x + 1, 0)
  init_chat_window (&inpw, w.ws_row-INP_W_LEN, 0);
  // end of GUI initialization
  state = Initialized;
  cn = net_new ();
  // init connection to the server
  if (conn_to_server__H () != 0)
    return -1;
  // try to login
  if (login__H () != 0)
    return -1;
  // save configuration to the default path
  save_config (default_config_path);
  // print chat rooms and ask user to select one
  select_chat__H ();
  
  wchar_t *buf = malloc (MAX_BUF * sizeof (wchar_t));
  memset (buf, 0, MAX_BUF * sizeof (wchar_t));
  
  while (cn.state == Connected)
    {
      cw_read (&inpw, buf, MAX_BUF);
      if (buf[0]=='\0')
        continue;
      if (buf[0]=='E' && buf[1]=='O' && buf[2]=='F')
        break;

      int len;
      if (state != Joined)
        { // send chat select packet
          net_wwrite (&cn, CHAT_SELECT, buf);
          const char *p = net_read (&cn, NULL);
          if (strncmp (p, "Chat doesn't exist", 18) == 0)
            SAFE_WRITE (cw_write_char (&cw, " ? chat not found - try again"));
          else if ((len = strlen (p)) != 0)
            {
              isJoined = true;
              cw_clear (&cw);
              SAFE_WRITE ({
                  char *dash_header = malloc (len + 1);

                  memset (dash_header, '-', len);
                  cw_vawrite_char (&cw, 6,
                            dash_header, "\n", p, "\n", dash_header, "\n");

                  free (dash_header);
                });
              state = Joined;
              wcharcpy (chatID, buf);
              cw_set_name (&cw, chatID);
            }
        }
      else
        { // send text packet
          if (net_wwrite (&cn, TEXT, buf) <= 0)
            RET_ERR ("Network Fault -- Exiting.", -1);
          SAFE_WRITE (cw_write_my_mess (&cw, buf));
        }
    }
  free (buf);
  return 0;
}

static inline int
READCHAT_loop_H (void *)
{
  int n;
  char *p;
  
  while (state != Joined) {};
  while (cn.state == Connected)
    { // text message
      p = net_read (&cn, &n);
      if (p == NULL)
        RET_ERR ("Network Fault -- Exiting.", -1);
      if (strlen (p) != 0)
        SAFE_WRITE (cw_write_char_mess (&cw, p));
    }

  return 0; // unreachable
}

static inline int
load_config_from_file (const char *path)
{
  int res = 0, r = 2;
  FILE *f;

  if (path == NULL || strlen (path) == 0)
      f = fopen (default_config_path, "r");
  else
    f = fopen (path, "r");
  
  if (f==NULL)
    RET_ERR ("Could Not Open the Config File -- Exiting.", -1);
  
  char *key = malloc (32);
  char *val = malloc (128);
  
  while (r == 2)
    {
      r = fscanf (f, "%32[^ ] %128[^\n]%*c", key, val);
      if (r < 0)
        RET_ERR ("Could Not Read the Config File -- Exiting.", -1);
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
        RET_ERR ("Parsing Config File Failed -- Exiting.", -1);
    }
  
  free (key);
  free (val);
  fclose (f);
  return res;
}

static inline int
save_config (const char *path)
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
get_arg (const char *flag, char *arg)
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
    RET_ERR ("unknown argument -- exiting.", 1);
  return 0;
}

static inline int
pars_args (int argc, char **argv)
{
  for (argc--, argv++; argc > 0; argc--, argv++)
    {
      if (!opt.EOO && argv[0][0] == '-')
        {
          if (argc == 1)
            RET_ERR ("Invalid Option Value -- Exiting.", -1);
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

static inline int
check_opts ()
{
  if (strlen (opt.server_addr) == 0)
    RET_ERR ("Invalid Server IP address -- Exiting.", -1);
  if (opt.server_port <= 0 || opt.server_port > (2<<16))
    RET_ERR ("Invalid Server Port Number -- Exiting.", -1);
  if (opt.username == NULL && opt.user_token == NULL)
    RET_ERR ("Either a Token or Username is Required -- Exiting", -1);

  return 0;
}

int
main (int argc, char **argv)
{
  __progname__ = argv[0];
  if (got_enough_space (w) != 0)
    {
      puts (ERR_MSG);
      return 1;
    }
  if (pars_args (argc, argv) != 0)
    {
      puts (ERR_MSG);
      Usage ();
      return 1;
    }

  // check opt
  if (check_opts () != 0)
    {
      puts (ERR_MSG);
      return 1;
    }
  
  thrd_t t;
  if (thrd_create (&t, READCHAT_loop_H, NULL) < 0)
    {
      puts ("Could Not Allocate Thread -- Exiting.");
      return 1;
    }
  MAIN_loop_H (NULL);

  thrd_detach (t);
  __exit ();

  if (ERR_MSG != NULL)
    puts (ERR_MSG);
  return 0;
}
