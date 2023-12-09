#ifndef NET__H__
#define NET__H__
#include "utils.h"

typedef enum {
  SIGNUP = 1,
  LOGIN_OUT,
  CHAT_SELECT,
  TEXT,
  WHOAMI
} Packet;

struct net_buf{
  char *buf;
  int cap;
  lock_t lock;
};
typedef struct {
  int sfd; // socket file descriptor
  int retry2conn; // max retry to reconnect to the server
  struct net_buf rbuf, wbuf;
} chat_net;

chat_net net_new();
int net_init(chat_net *, const char *, int);
int net_write(chat_net *, Packet, const char *, int);
int net_wwrite(chat_net *, Packet, const wchar_t *);
char * net_read(chat_net *, int *);
void net_end(chat_net *);

#endif
