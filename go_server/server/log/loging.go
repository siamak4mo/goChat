package log

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

var (
	loglevel uint
)

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

func New(cfg config.Config){
	loglevel = cfg.Log.LogLevel
}

func Logf(level Lev, format string, args ...any){
	if level >= Lev(loglevel) {
		fmt.Print(level2str(level))
		fmt.Printf(format, args...)
	}
}
