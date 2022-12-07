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

func SetLevel(lvl any) {
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

func Println(level Level, v ...any) {
	if DEBUG <= level && level <= CRITICAL && l.lvl <= level {
		l.Println(append([]any{fmt.Sprint(level) + ":"}, v...)...)
	}
}

func Printf(level Level, format string, v ...any) {
	if DEBUG <= level && level <= CRITICAL && l.lvl <= level {
		l.Printf("%s: "+format, append([]any{level}, v...)...)
	}
}

func Debug(v ...any) {
	Println(DEBUG, v...)
}

func Debugf(format string, v ...any) {
	Printf(DEBUG, format, v...)
}

func Info(v ...any) {
	Println(INFO, v...)
}

func Infof(format string, v ...any) {
	Printf(INFO, format, v...)
}

func Error(v ...any) {
	Println(ERROR, v...)
}

func Errorf(format string, v ...any) {
	Printf(ERROR, format, v...)
}

func Critical(v ...any) {
	Println(CRITICAL, v...)
}

func Criticalf(format string, v ...any) {
	Printf(CRITICAL, format, v...)
}
