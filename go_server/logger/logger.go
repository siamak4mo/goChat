package logger

import (
	"fmt"
	"server/config"
	"time"
)

type (
	Level  uint // log level
	Logf   func(format string, args ...any)
	LogExt func(*LogWriter) // log extension
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
	module    string
	log_level uint
}

type LogWriter struct {
	LogLevel Level
	Time     time.Time
	Module   string
}

func New(cfg config.Config, module_name string) *Log {
	return &Log{
		log_level: cfg.Log.LogLevel,
		module:    module_name,
	}
}

// internal function
// prints: `module name | epoch time [level]`
// and then normal printf with format and args
func (lw *LogWriter) Flushf(format string, args ...any) {
	lw.Time = time.Now()
	fmt.Printf("%s| %v %s", lw.Module, lw.Time.Unix(), lev_label[lw.LogLevel])
	fmt.Printf(format, args...)
}

// helper function maker
func (l *Log) funMaker(level Level, extensions ...LogExt) Logf {
	return func(format string, args ...any) {
		if level >= Level(l.log_level) {
			lw := &LogWriter{
				LogLevel: level,
				Time:     time.Now(),
				Module:   l.module,
			}
			for _, fun := range extensions {
				fun(lw)
			}
			lw.Flushf(format, args)
		}
	}
}

func (l *Log) Debugf(extensions ...LogExt) Logf {
	return l.funMaker(L_Debug, extensions...)
}

func (l *Log) Infof(extensions ...LogExt) Logf {
	return l.funMaker(L_Info, extensions...)
}

func (l *Log) Warnf() Logf {
	return l.funMaker(L_Warning)
}

func (l *Log) Errorf() Logf {
	return l.funMaker(L_Error)
}

func (l *Log) Panicf() Logf {
	return l.funMaker(L_Panic)
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

/* Extensions               */
/* naming format: LogEX_xxx */
// add a sub module (prints: `Module.sub_module`)
func LogEX_SetSubModule(sub_module string) LogExt {
	return func(l *LogWriter) {
		l.Module = fmt.Sprintf("%s.%s", l.Module, sub_module)
	}
}

// update current time
func LogEX_UpdateTime() LogExt {
	return func(l *LogWriter) {
		l.Time = time.Now()
	}
}

// set module
func LogEX_SetModule(module string) LogExt {
	return func(l *LogWriter) {
		l.Module = module
	}
}

// set time
func LogEX_SetTime(t time.Time) LogExt {
	return func(l *LogWriter) {
		l.Time = t
	}
}
