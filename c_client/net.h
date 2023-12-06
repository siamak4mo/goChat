#ifndef NET__H__
#define NET__H__

typedef enum {
  SIGNUP = 1,
  LOGIN_OUT,
  CHAT_SELECT,
  TEXT
} Packet;

int net_init(const char *, int);
void net_write(Packet, const char *, int);
char * net_read(int *);
void net_end();

#endif
