#include <ncurses.h>
#include <stdlib.h>
#include <string.h>
#include "chat_window.h"

#define NL 3
#define NPAD 2

chatw
mk_chatw(int row, int col, bool boxed)
{
  return (chatw){.row=row, .col=col, .box=boxed, .w=NULL};
}

static inline void
set_name(chatw *cw)
{
  if (strlen (cw->name) == 0 || (cw->col)/3 < 6)
    return;
  
  size_t maxlen = (cw->col)/NL - NL;
  if (strlen (cw->name) <= maxlen)
    mvwprintw (cw->w, 0, NPAD, cw->name);
  else
    {
      char *name_cpy = malloc (maxlen + NL);
      strncpy (name_cpy, cw->name, maxlen);
      memset (name_cpy + maxlen, '.', NL - 1);
      name_cpy[maxlen + NL - 1]='\0';
      
      mvwprintw (cw->w, 0, NPAD, name_cpy);
    }
}

static inline void
redrawbox(chatw *cw)
{
  werase (cw->w);
  box (cw->w, 0, 0);
  set_name (cw);
  wrefresh (cw->w);
}

void
init_chat_window(chatw *cw, int x, int y)
{
  if (cw->box)
    cw->padding = 1;
  else cw->padding = 0;
  
  cw->w = newwin (cw->row, cw->col, x,y);
  cw->line_c = cw->padding;

  refresh ();

  if (cw->box)
    redrawbox (cw);
}

static inline void
lift_up(chatw *cw, int n)
{
  int i,j;

  if (n >= cw->row)
    {
      werase (cw->w);
      redrawbox (cw);
    }
  else
    {
      for (i=cw->padding; i < cw->row - n - 1; ++i)
        for (j = cw->padding; j < cw->col - cw->padding; ++j)
          {
            int c = mvwinch (cw->w, i+n, j);
            mvwaddch (cw->w, i, j, c);
          }

      for (i = cw->col - n; i<cw->col; ++i)
        for (j = cw->padding; j < cw->col - cw->padding; ++j)
          mvwaddch (cw->w, i, j, ' ');
    }
}

static inline void
lift_up1(chatw *cw)
{
  int i, j;

  for (i=cw->padding; i < cw->row - cw->padding -1; ++i)
    for (j = cw->padding; j < cw->col - cw->padding; ++j)
      {
        wchar_t c = mvwinch (cw->w, i+1, j);
        mvwaddch (cw->w, i, j, c);
      }

  for (i = cw->col - 1; i<cw->col; ++i)
    for (j = cw->padding; j < cw->col - cw->padding; ++j)
      mvwaddch (cw->w, i, j, ' ');
}

void
cw_write_char(chatw *cw, const char *buf)
{
  int i;
  
  for (i = cw->padding; *buf != '\0'; ++buf)
    {
      if (cw->line_c >= cw->row - cw->padding)
        {
          lift_up1 (cw);
          cw->line_c = cw->row - cw->padding - 1;
          for (; i<cw->col-cw->padding; ++i)
            mvwaddch (cw->w, cw->line_c, i, ' ');
          i = cw->padding;
        }

      if (*buf != '\n')
        {
          mvwaddch (cw->w, cw->line_c, i++, *buf);
          if (i==cw->col-cw->padding)
            {
              i = cw->padding;
              if (*(buf+1) != '\0')
                cw->line_c++;
            }
        }
      else
        {
          cw->line_c++;
          i = cw->padding;
        }
      
    }
  if (cw->line_c < cw->row)
    cw->line_c++;
  wrefresh (cw->w);
} 

void
cw_write(chatw *cw, const wchar_t *buf)
{
  int i;
  
  for (i = cw->padding; *buf != '\0'; ++buf)
    {
      if (cw->line_c >= cw->row - cw->padding)
        {
          lift_up1 (cw);
          cw->line_c = cw->row - cw->padding - 1;
          for (; i<cw->col-cw->padding; ++i)
            mvwaddch (cw->w, cw->line_c, i, ' ');
          i = cw->padding;
        }

      if (*buf != '\n')
        {
          mvwaddch (cw->w, cw->line_c, i++, *buf);
          if (i==cw->col-cw->padding)
            {
              i = cw->padding;
              if (*(buf+1) != '\0')
                cw->line_c++;
            }
        }
      else
        {
          cw->line_c++;
          i = cw->padding;
        }
      
    }
  if (cw->line_c < cw->row)
    cw->line_c++;
  wrefresh (cw->w);
}

static inline void
reset_read(chatw *cw)
{
  cw->line_c = cw->padding;
  werase (cw->w);
  refresh ();
  if (cw->box)
    {
      box (cw->w, 0, 0);
      set_name (cw);
    }
  wrefresh (cw->w);
}

int
cw_read(chatw *cw, wchar_t *result, int maxlen)
{
  int rw = 0;
  int i;

  i = cw->padding;
  reset_read(cw);
  
  while (rw != maxlen)
    {
      if (cw->line_c >= cw->row - 1)
        {
          lift_up1 (cw);
          cw->line_c = cw->row - cw->padding - 1;
          for (; i<cw->col-1; ++i)
            mvwaddch (cw->w, cw->line_c, i, ' ');
          i = cw->padding;
        }

      if (i != cw->col - cw->padding)
        {
          noecho ();
          result[rw] = mvwgetch (cw->w, cw->line_c, i);
          if (result[rw]==127)
            {
              if (i>cw->padding)
                {
                  mvwaddch (cw->w, cw->line_c, i-1, ' ');
                  result[rw] = 0;
                  result[rw-1] = 0;
                  rw -= 2;
                  i -= 1;
                }
              else
                {
                  if (cw->line_c > cw->padding)
                    {
                      cw->line_c--;
                      i = cw->col - cw->padding-1;
                      
                      mvwaddch (cw->w, cw->line_c, i, ' ');
                      result[rw-1] = 0;
                      rw -= 2;
                    }
                  else
                    {
                      if (rw>1)
                        {
                          for (int k=cw->col - cw->padding;
                               k>cw->padding; --k)
                            mvwaddch (cw->w, cw->line_c, cw->col-k,
                                      result[rw + cw->padding - k]);
                          i = cw->col - cw->padding-1;
                          mvwaddch (cw->w, cw->line_c, i, ' ');
                          result[rw-1] = 0;
                          rw -= 2;
                        }
                      else
                        rw=-1; // to skip rw++ at the end
                    }
                }
            }
          else if (result[rw]=='\n')
            {
              result[rw] = 0;
              return rw;
            }
          else
            mvwaddch (cw->w, cw->line_c, i++, result[rw]);
          
          if (i==cw->col-cw->padding)
            {
              i = cw->padding;
              cw->line_c++;
            }
        }
      else
        {
          cw->line_c++;
          i = cw->padding;
        }
      
      rw++;
    }
  if (cw->line_c < cw->row)
    cw->line_c++;
  wrefresh (cw->w);
  
  return rw;
}
