package serlog

import (
	"fmt"
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
func (l Log) logf(level Lev, format string, args ...any) {
	if level >= Lev(l.log_level) {
		l.time = time.Now()
		fmt.Printf("%s| %v %s", l.module, l.time.Unix(), lev_label[level])
		fmt.Printf(format, args...)
	}
}

func (l Log) Debugf(format string, args ...any) {
	l.logf(Debug, format, args...)
}
func (l Log) Infof(format string, args ...any) {
	l.logf(Info, format, args...)
}
func (l Log) Warnf(format string, args ...any) {
	l.logf(Warning, format, args...)
}

func (l Log) Errorf(format string, args ...any) {
	l.logf(Error, format, args...)
}
func (l Log) Panicf(format string, args ...any) {
	l.logf(Panic, format, args...)
}

func (l Log) Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}

// prints: `module name | `
// and then normal printf with format and args
func (l Log) Pprintf(format string, args ...any) {
	fmt.Printf("%s| ", l.module)
	fmt.Printf(format, args...)
}
