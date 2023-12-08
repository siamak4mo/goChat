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


// for loop for wchar_t
// @_ptr should be a char pointer
#define WCHAR4(_ptr, X)                         \
  for (_ptr = (char*) X;                        \
       !(_ptr[0] == 0 && _ptr[1] == 0 &&        \
         _ptr[2] == 0 && _ptr[3] == 0); _ptr += 4)

// from char pointer @X to wchar_t @_ptr
// @X should be a char pointer and @_ptr wchar_t
#define WCHAR2(_wch, X)                         \
  for (char *__p = X;                           \
       ({(_wch) = *__p; *__p != '\0';}); __p++)

#define wcharcpy(dest, src)                     \
  wchar_t *X = src;                             \
  char *_ptr, *p = dest;                        \
  WCHAR4(_ptr, X) *(p++) = *_ptr;

#endif
