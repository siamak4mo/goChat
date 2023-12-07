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

static int sfd = INVALID_SOCKET; // socket fd
static char *buf;
static int lock = 0;
static int max_retry = 4;

#define LOCK() lock = 1
#define UNLOCK() lock = 0

#define WAIT_LOCK() do {                        \
    while (lock) {};                            \
    LOCK();            } while (0)


int
net_init(const char *addr, int port)
{
  int res;
  struct sockaddr_in ss_in;

  ss_in.sin_family = AF_INET;
  ss_in.sin_addr.s_addr = inet_addr(addr);
  ss_in.sin_port = htons(port);

  while (max_retry-- != 0)
    {
      sfd = socket (AF_INET, SOCK_STREAM, 0);
      res = connect (sfd,
                         (struct sockaddr *) &ss_in,
                         sizeof(ss_in));
      if (res==0)
        {
          buf = malloc (MAX_TR_SIZE);
          return 0;
        }
      else
        sleep (1);
    }  

  return -1;
}

static inline void
set_packet_type(Packet type)
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
net_write(Packet type, const char *body, int len)
{
  WAIT_LOCK();
  
  set_packet_type (type);

  if (len > 0)
    memcpy (buf + PAC_PAD, body, len);
  
  write (sfd, buf, len + PAC_PAD);

  UNLOCK();
}

const char *
net_read(int *len)
{
  WAIT_LOCK();
  
  int _r = read (sfd, buf, MAX_TR_SIZE);
  
  buf[_r] = 0;
  if (_r > 0)
    buf[_r-1] = 0; // remove '\n'
  
  if (len != NULL)
    *len = _r;

  UNLOCK();
  return buf;
}

void
net_end()
{ 
  WAIT_LOCK();

  close (sfd);
  sfd = INVALID_SOCKET;
  free (buf);

  UNLOCK();
}
