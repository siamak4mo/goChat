package serlog

import (
	"fmt"
	"server/config"
	"time"
)

type (
	Lev       uint
	Logf func(format string, args ...any)
)

const (
	Debug Lev = iota + 1
	Info
	Warning
	Error
	Panic
)

var (
	lev_label = map[Lev]string{
		Debug:   "[DEBUG] ",
		Info:    "[INFO] ",
		Warning: "[WARNING] ",
		Error:   "ERROR -- ",
		Panic:   " ** PANIC ** ",
	}
)

type Log struct {
	time      time.Time
	module    string
	log_level uint
}

func New(cfg config.Config, module_name string) *Log {
	return &Log{
		log_level: cfg.Log.LogLevel,
		time:      time.Now(),
		module:    module_name,
	}
}

// internal function
// prints: `module name | epoch time [level]`
// and then normal printf with format and args
func flushf(l *Log, level Lev, format string, args ...any) {
	if level >= Lev(l.log_level) {
		l.time = time.Now()
		fmt.Printf("%s| %v %s", l.module, l.time.Unix(), lev_label[level])
		fmt.Printf(format, args...)
	}
}

func (l *Log) Debugf() Logf {
	return func(format string, args ...any) {
		flushf(l, Debug, format, args)
	}
}
func (l *Log) Infof() Logf {
	return func(format string, args ...any) {
		flushf(l, Info, format, args)
	}
}
func (l *Log) Warnf() Logf {
	return func(format string, args ...any) {
		flushf(l, Warning, format, args)
	}

}

func (l *Log) Errorf() Logf {
	return func(format string, args ...any) {
		flushf(l, Error, format, args)
	}

}
func (l *Log) Panicf() Logf {
	return func(format string, args ...any) {
		flushf(l, Panic, format, args)
	}

}

func (l *Log) Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}

// prints: `module name | `
// and then normal printf with format and args
func (l *Log) Pprintf(format string, args ...any) {
	fmt.Printf("%s| ", l.module)
	fmt.Printf(format, args...)
}
