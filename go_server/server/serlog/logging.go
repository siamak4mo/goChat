package serlog

import (
	"fmt"
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
	switch level{
		case Debug: return "[Debug] "
		case Info: return "[INFO] "
		case Warning: return "[Warning] "
		case Error: return "ERROR -- "
		case Panic: return "*PANIC* -- "
	}
	return ""
}

func New(cfg config.Config) *Log{
	return &Log{
		LogLevel: cfg.Log.LogLevel,
	}
}

func (l Log) Logf(level Lev, format string, args ...any){
	if level >= Lev(l.LogLevel) {
		fmt.Print(level2str(level))
		fmt.Printf(format, args...)
	}
}
