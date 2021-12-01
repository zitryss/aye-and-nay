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
	ldisabled Level = iota
	Ldebug          // debug
	Linfo           // info
	Lerror          // error
	Lcritical       // critical
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
	l.SetPrefix(prefix)
}

func SetFlags(flag int) {
	l.SetFlags(flag)
}

func SetLevel(lvl interface{}) {
	oldLevel := l.lvl
	newLevel := ldisabled
	switch v := lvl.(type) {
	case Level:
		if Ldebug <= v && v <= Lcritical {
			newLevel = v
		} else {
			newLevel = oldLevel
		}
	case string:
		v = strings.ToLower(v)
		switch v {
		case "debug":
			newLevel = Ldebug
		case "info":
			newLevel = Linfo
		case "error":
			newLevel = Lerror
		case "critical":
			newLevel = Lcritical
		default:
			newLevel = oldLevel
		}
	default:
		newLevel = oldLevel
	}
	l.lvl = newLevel
}

func Println(level Level, v ...interface{}) {
	if Ldebug <= level && level <= Lcritical && l.lvl <= level {
		l.Println(append([]interface{}{fmt.Sprint(level) + ":"}, v...)...)
	}
}

func Printf(level Level, format string, v ...interface{}) {
	if Ldebug <= level && level <= Lcritical && l.lvl <= level {
		l.Printf("%s: "+format, append([]interface{}{level}, v...)...)
	}
}

func Debug(v ...interface{}) {
	Println(Ldebug, v...)
}

func Debugf(format string, v ...interface{}) {
	Printf(Ldebug, format, v...)
}

func Info(v ...interface{}) {
	Println(Linfo, v...)
}

func Infof(format string, v ...interface{}) {
	Printf(Linfo, format, v...)
}

func Error(v ...interface{}) {
	Println(Lerror, v...)
}

func Errorf(format string, v ...interface{}) {
	Printf(Lerror, format, v...)
}

func Critical(v ...interface{}) {
	Println(Lcritical, v...)
}

func Criticalf(format string, v ...interface{}) {
	Printf(Lcritical, format, v...)
}
