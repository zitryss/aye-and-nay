package log

import (
	"io"
	"io/ioutil"
	"log"
	"strings"
)

type level int

const (
	ldisabled level = iota
	Ldebug
	Linfo
	Lerror
	Lcritical
)

var (
	l = logger{Logger: log.New(ioutil.Discard, "", log.LstdFlags)}
)

type logger struct {
	*log.Logger
	level
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

func SetLevel(lev interface{}) {
	oldLevel := l.level
	newLevel := ldisabled
	switch v := lev.(type) {
	case level:
		newLevel = v
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
	l.level = newLevel
}

func Debug(v ...interface{}) {
	if l.level <= Ldebug {
		l.Println(append([]interface{}{"debug:"}, v...)...)
	}
}

func Debugf(format string, v ...interface{}) {
	if l.level <= Ldebug {
		l.Printf("debug: "+format, v...)
	}
}

func Info(v ...interface{}) {
	if l.level <= Linfo {
		l.Println(append([]interface{}{"info:"}, v...)...)
	}
}

func Infof(format string, v ...interface{}) {
	if l.level <= Linfo {
		l.Printf("info: "+format, v...)
	}
}

func Error(v ...interface{}) {
	if l.level <= Lerror {
		l.Println(append([]interface{}{"error:"}, v...)...)
	}
}

func Errorf(format string, v ...interface{}) {
	if l.level <= Lerror {
		l.Printf("error: "+format, v...)
	}
}

func Critical(v ...interface{}) {
	if l.level <= Lcritical {
		l.Println(append([]interface{}{"critical:"}, v...)...)
	}
}

func Criticalf(format string, v ...interface{}) {
	if l.level <= Lcritical {
		l.Printf("critical: "+format, v...)
	}
}
