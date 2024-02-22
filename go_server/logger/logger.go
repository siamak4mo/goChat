package serlog

import (
	"fmt"
	"server/config"
	"time"
)

type (
	Level  uint // log level
	Logf   func(format string, args ...any)
	LogExt func(*Log) // log extension
)

const (
	L_Debug Level = iota + 1
	L_Info
	L_Warning
	L_Error
	L_Panic
)

var (
	lev_label = map[Level]string{
		L_Debug:   "[DEBUG] ",
		L_Info:    "[INFO] ",
		L_Warning: "[WARNING] ",
		L_Error:   "ERROR -- ",
		L_Panic:   " ** PANIC ** ",
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
func flushf(l *Log, level Level, format string, args ...any) {
	if level >= Level(l.log_level) {
		l.time = time.Now()
		fmt.Printf("%s| %v %s", l.module, l.time.Unix(), lev_label[level])
		fmt.Printf(format, args...)
	}
}

// this is a LogExt (extension) example
func LogUpdateTime(l *Log) {
	l.time = time.Now()
}

func (l *Log) Debugf(extensions ...LogExt) Logf {
	for _, fun := range extensions {
		fun(l)
	}
	return func(format string, args ...any) {
		flushf(l, L_Debug, format, args)
	}
}

func (l *Log) Infof(extensions ...LogExt) Logf {
	for _, fun := range extensions {
		fun(l)
	}
	return func(format string, args ...any) {
		flushf(l, L_Info, format, args)
	}
}

func (l *Log) Warnf() Logf {
	return func(format string, args ...any) {
		flushf(l, L_Warning, format, args)
	}
}

func (l *Log) Errorf() Logf {
	return func(format string, args ...any) {
		flushf(l, L_Error, format, args)
	}
}

func (l *Log) Panicf() Logf {
	return func(format string, args ...any) {
		flushf(l, L_Panic, format, args)
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
