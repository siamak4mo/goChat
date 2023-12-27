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

static inline void
net_malloc(struct net_buf *netb)
{
  if (netb->cap > 0)
    netb->buf = malloc (netb->cap);
}

static inline void
net_free(struct net_buf *netb)
{
  if (netb->buf != NULL && netb->cap > 0)
    {
      free (netb->buf);
      netb->cap = -1;
    }
}

chat_net
net_new()
{
  struct net_buf netb_R = {
    .buf=NULL,
    .cap=MAX_TR_SIZE,
  };
  struct net_buf netb_W = {
    .buf=NULL,
    .cap=MAX_TR_SIZE,
  };

  chat_net cn = {
    .sfd=INVALID_SOCKET,
    .retry2conn=4,
    .rbuf=netb_R,
    .wbuf=netb_W,
    .state=Unestablished
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
          net_malloc (&(cn->rbuf));
          net_malloc (&(cn->wbuf));
          cn->state = Connected;
          return 0;
        }
      else
        sleep (1);
    }  

  cn->state = Disconnected;
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

int
net_write(chat_net *cn, Packet type,
          const char *body, int len)
{
  char *buf;

  if (cn->state != Connected)
    return -1;
  
  buf = (cn->wbuf).buf;
  set_packet_type (buf, type);

  if (len > (cn->wbuf).cap - PAC_PAD)
    len = (cn->wbuf).cap - PAC_PAD - 1;
  
  if (len > 0)
    memcpy (buf + PAC_PAD, body, len);
  buf[len + PAC_PAD] = '\n';
  
  ssize_t res = write (cn->sfd, buf, len + PAC_PAD + 1);
  if (res < 0)
    cn->state = Disconnected;
  return res;
}

int
net_wwrite(chat_net *cn, Packet type, const wchar_t *body)
{
  char *buf;
  char *p;
  int len = 0;

  if (cn->state != Connected)
    return -1;
  
  buf = (cn->wbuf).buf;
  set_packet_type (buf, type);

  WCHAR4(p, body)
    {
      buf[len + PAC_PAD] = *p;
      if (++len >= (cn->wbuf).cap)
        break;
    }
  buf[len + PAC_PAD] = '\n';
  
  ssize_t res = write (cn->sfd, buf, len + PAC_PAD + 1);
  if (res < 0)
    cn->state = Disconnected;
  return res;
}

char *
net_read(chat_net *cn, int *len)
{
  char *buf;
  
  if (cn->state != Connected)
    return "not conected";
  
  buf = (cn->rbuf).buf;
  ssize_t _r = read (cn->sfd, buf, (cn->rbuf).cap);
  if (_r <= 0)
    {
      cn->state = Disconnected;
      if (len != NULL)
        *len = _r;
      return NULL;
    }
  
  buf[_r] = 0;
  if (_r > 0)
    buf[_r-1] = 0; // remove '\n'
  
  if (len != NULL)
    *len = _r;

  return buf;
}

void
net_end(chat_net *cn)
{ 
  close (cn->sfd);
  cn->sfd = INVALID_SOCKET;
  net_free (&(cn->rbuf));
  net_free (&(cn->wbuf));
}
