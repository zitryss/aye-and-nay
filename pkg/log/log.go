//go:generate $GOPATH/bin/stringer -type=Level -linecomment
package log

import (
	"fmt"
	"io"
	"log"
	"strings"
)

type Level int

const (
	disabled Level = iota
	DEBUG          // debug
	INFO           // info
	ERROR          // error
	CRITICAL       // critical
)

var (
	l = logger{Logger: log.New(io.Discard, "", log.LstdFlags)}
)

type logger struct {
	*log.Logger
	lvl Level
}

func SetOutput(w io.Writer) {
	l.SetOutput(w)
}

func SetPrefix(prefix string) {
	if len(prefix) > 0 {
		l.SetPrefix(prefix + ": ")
	}
}

func SetFlags(flag int) {
	l.SetFlags(flag)
}

func SetLevel(lvl interface{}) {
	oldLevel := l.lvl
	newLevel := disabled
	switch v := lvl.(type) {
	case Level:
		if DEBUG <= v && v <= CRITICAL {
			newLevel = v
		} else {
			newLevel = oldLevel
		}
	case string:
		v = strings.ToLower(v)
		switch v {
		case "debug":
			newLevel = DEBUG
		case "info":
			newLevel = INFO
		case "error":
			newLevel = ERROR
		case "critical":
			newLevel = CRITICAL
		default:
			newLevel = oldLevel
		}
	default:
		newLevel = oldLevel
	}
	l.lvl = newLevel
}

func Println(level Level, v ...interface{}) {
	if DEBUG <= level && level <= CRITICAL && l.lvl <= level {
		l.Println(append([]interface{}{fmt.Sprint(level) + ":"}, v...)...)
	}
}

func Printf(level Level, format string, v ...interface{}) {
	if DEBUG <= level && level <= CRITICAL && l.lvl <= level {
		l.Printf("%s: "+format, append([]interface{}{level}, v...)...)
	}
}

func Debug(v ...interface{}) {
	Println(DEBUG, v...)
}

func Debugf(format string, v ...interface{}) {
	Printf(DEBUG, format, v...)
}

func Info(v ...interface{}) {
	Println(INFO, v...)
}

func Infof(format string, v ...interface{}) {
	Printf(INFO, format, v...)
}

func Error(v ...interface{}) {
	Println(ERROR, v...)
}

func Errorf(format string, v ...interface{}) {
	Printf(ERROR, format, v...)
}

func Critical(v ...interface{}) {
	Println(CRITICAL, v...)
}

func Criticalf(format string, v ...interface{}) {
	Printf(CRITICAL, format, v...)
}
