#include <unistd.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <arpa/inet.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include "net.h"

#define MAX_TR_SIZE 512
#define PAC_PAD 2
#define INVALID_SOCKET -1

#define _GET_LOCK(cn) ((cn)->nbuf).lock
#define LOCK(cn) _LOCK(_GET_LOCK(cn)))
#define UNLOCK(cn) _UNLOCK(_GET_LOCK(cn))
#define WAIT_LOCK(cn) _WAIT_LOCK(_GET_LOCK(cn))


static inline void
net_malloc(struct net_buf *netb)
{
  while (netb->lock == locked) {};
  netb->lock = locked;
  
  if (netb->cap > 0)
    netb->buf = malloc (netb->cap);
  
  netb->lock = unlocked;
}

static inline void
net_free(struct net_buf *netb)
{
  while (netb->lock == locked) {};
  netb->lock = locked;
  
  free (netb->buf);
  netb->cap = -1;

  netb->lock = unlocked;
}

chat_net
net_new()
{
  struct net_buf netb = {
    .buf=NULL,
    .cap=MAX_TR_SIZE,
    .lock=unlocked
  };

  chat_net cn = {
    .sfd=INVALID_SOCKET,
    .retry2conn=4,
    .nbuf=netb
  };

  return cn;
}


int
net_init(chat_net *cn, const char *addr, int port)
{
  int res;
  struct sockaddr_in ss_in;

  ss_in.sin_family = AF_INET;
  ss_in.sin_addr.s_addr = inet_addr(addr);
  ss_in.sin_port = htons(port);

  while (cn->retry2conn-- != 0)
    {
      cn->sfd = socket (AF_INET, SOCK_STREAM, 0);
      res = connect (cn->sfd,
                     (struct sockaddr *) &ss_in,
                     sizeof(ss_in));
      if (res==0)
        {
          net_malloc (&(cn->nbuf));
          return 0;
        }
      else
        sleep (1);
    }  

  return -1;
}

static inline void
set_packet_type(char *buf, Packet type)
{
  switch(type)
    {
    case SIGNUP:
      memcpy (buf, "S ", PAC_PAD);
      break;
    case LOGIN_OUT:
      memcpy (buf, "L ", PAC_PAD);
      break;
    case CHAT_SELECT:
      memcpy (buf, "C ", PAC_PAD);
      break;
    case TEXT:
      memcpy (buf, "T ", PAC_PAD);
      break;
    case WHOAMI:
      memcpy (buf, "W ", PAC_PAD);
      break;
    }
}

void
net_write(chat_net *cn, Packet type,
          const char *body, int len)
{
  char *buf;
  WAIT_LOCK(cn);
  
  buf = (cn->nbuf).buf;
  set_packet_type (buf, type);

  if (len > (cn->nbuf).cap - PAC_PAD)
    len = (cn->nbuf).cap - PAC_PAD - 1;
  
  if (len > 0)
    memcpy (buf + PAC_PAD, body, len);
  buf[len + PAC_PAD] = '\n';
  
  write (cn->sfd, buf, len + PAC_PAD + 1);

  UNLOCK(cn);
}

const char *
net_read(chat_net *cn, int *len)
{
  WAIT_LOCK(cn);
  char *buf = (cn->nbuf).buf;
  
  int _r = read (cn->sfd, buf,
                 (cn->nbuf).cap);
  
  buf[_r] = 0;
  if (_r > 0)
    buf[_r-1] = 0; // remove '\n'
  
  if (len != NULL)
    *len = _r;

  UNLOCK(cn);
  return buf;
}

void
net_end(chat_net *cn)
{ 
  WAIT_LOCK(cn);

  close (cn->sfd);
  cn->sfd = INVALID_SOCKET;
  net_free (&(cn->nbuf));

  UNLOCK(cn);
}
