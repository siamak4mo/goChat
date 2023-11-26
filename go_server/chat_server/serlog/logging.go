package serlog

import (
	"fmt"
	"os"
	"server/server/config"
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

type Log struct {
	Time     time.Time
	LogLevel uint
}

func level2str(level Lev) string {
	switch level {
	case Debug:
		return "[Debug] "
	case Info:
		return "[INFO] "
	case Warning:
		return "[Warning] "
	case Error:
		return "ERROR -- "
	case Panic:
		return " ** PANIC **  "
	}
	return ""
}

func New(cfg config.Config) *Log {
	return &Log{
		LogLevel: cfg.Log.LogLevel,
		Time: time.Now(),
	}
}

func (l Log) logf(level Lev, fun func(), format string, args ...any) {
	defer fun()

	if level >= Lev(l.LogLevel) {
		l.Time = time.Now()
		fmt.Printf("%v %s", l.Time.Unix(),level2str(level))
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
