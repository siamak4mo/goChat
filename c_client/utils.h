#ifndef UTILES_DOT_H__
#define UTILES_DOT_H__
#include <stdbool.h>

#define locked true
#define unlocked false

typedef bool lock_t;

#define _LOCK(x) x = locked
#define _UNLOCK(x) x = unlocked

#define _WAIT_LOCK(x) do {                      \
  while (x) {};                                 \
  _LOCK(x);              } while (0)


#endif
