package serlog

import (
	"fmt"
	"os"
	"server/config"
	"time"
)

type Lev uint

const (
	Debug Lev = iota + 1
	Info
	Warning
	Error
	Panic
)

var (
	level2str = map[Lev]string{
		Debug:   "[DEBUG] ",
		Info:    "[INFO] ",
		Warning: "[WARNING] ",
		Error:   "ERROR -- ",
		Panic:   " ** PANIC ** ",
	}
)

type Log struct {
	Time     time.Time
	Module   string
	LogLevel uint
}

func New(cfg config.Config, module_name string) *Log {
	return &Log{
		LogLevel: cfg.Log.LogLevel,
		Time:     time.Now(),
		Module:   module_name,
	}
}

func (l Log) logf(level Lev, fun func(), format string, args ...any) {
	defer fun()

	if level >= Lev(l.LogLevel) {
		l.Time = time.Now()
		fmt.Printf("%s| %v %s", l.Module, l.Time.Unix(), level2str[level])
		fmt.Printf(format, args...)
	}
}

func Nop() {}
func ExitErr() {
	os.Exit(1)
}

func (l Log) Debugf(format string, args ...any) {
	l.logf(Debug, Nop, format, args...)
}
func (l Log) Infof(format string, args ...any) {
	l.logf(Info, Nop, format, args...)
}
func (l Log) Warnf(format string, args ...any) {
	l.logf(Warning, Nop, format, args...)
}

func (l Log) Errorf(format string, fun func(), args ...any) {
	if fun != nil {
		l.logf(Error, fun, format, args...)
	} else {
		l.logf(Error, ExitErr, format, args...)
	}
}
func (l Log) Panicf(format string, fun func(), args ...any) {
	if fun != nil {
		l.logf(Panic, fun, format, args...)
	} else {
		l.logf(Panic, ExitErr, format, args...)
	}
}

func (l Log) Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (l Log) Pprintf(format string, args ...any) {
	fmt.Printf("%s| ", l.Module)
	fmt.Printf(format, args...)
}
