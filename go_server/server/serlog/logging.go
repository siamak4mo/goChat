package serlog

import (
	"fmt"
	"os"
	"server/server/config"
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
		return "*PANIC* -- "
	}
	return ""
}

func New(cfg config.Config) *Log {
	return &Log{
		LogLevel: cfg.Log.LogLevel,
	}
}

func (l Log) logf(level Lev, def func(), format string, args ...any) {
	defer def()

	if level >= Lev(l.LogLevel) {
		fmt.Print(level2str(level))
		fmt.Printf(format, args...)
	}
}

func nop() {}
func exit() {
	os.Exit(1)
}

func (l Log) Debugf(format string, args ...any) {
	l.logf(Debug, nop, format, args...)
}
func (l Log) Infof(format string, args ...any) {
	l.logf(Info, nop, format, args...)
}
func (l Log) Warnf(format string, args ...any) {
	l.logf(Warning, nop, format, args...)
}

func (l Log) Errorf(format string, fun func(), args ...any) {
	if fun != nil {
		l.logf(Error, fun, format, args...)
	} else {
		l.logf(Error, exit, format, args...)
	}
}
func (l Log) Panicf(format string, fun func(), args ...any) {
	if fun != nil {
		l.logf(Panic, fun, format, args...)
	} else {
		l.logf(Panic, exit, format, args...)
	}
}
