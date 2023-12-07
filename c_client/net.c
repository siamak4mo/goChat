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
  free (netb->buf);
  netb->cap = -1;
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
    .wbuf=netb_W
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

  buf = (cn->wbuf).buf;
  set_packet_type (buf, type);

  if (len > (cn->wbuf).cap - PAC_PAD)
    len = (cn->wbuf).cap - PAC_PAD - 1;
  
  if (len > 0)
    memcpy (buf + PAC_PAD, body, len);
  buf[len + PAC_PAD] = '\n';
  
  write (cn->sfd, buf, len + PAC_PAD + 1);
}

void
net_wwrite(chat_net *cn, Packet type, const wchar_t *body)
{
  char *buf;
  char *p;
  int len = 0;

  p = (char*)body;
  buf = (cn->wbuf).buf;
  set_packet_type (buf, type);

  while (len < (cn->wbuf).cap &&
         !(p[0] == 0 && p[1] == 0 && p[2] == 0 && p[3] == 0))
    {
      buf[len + PAC_PAD] = *p;
      len++;
      p += 4;
    }
  buf[len + PAC_PAD] = '\n';
  
  write (cn->sfd, buf, len + PAC_PAD + 1);
}

const char *
net_read(chat_net *cn, int *len)
{
  char *buf = (cn->rbuf).buf;
  int _r = read (cn->sfd, buf, (cn->rbuf).cap);
  
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
