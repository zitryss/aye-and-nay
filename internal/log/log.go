package log

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zitryss/aye-and-nay/internal/requestid"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

var (
	l logger
)

type logger struct {
	*zap.SugaredLogger
	lvl    zapcore.Level
	prefix string
}

func New(lvl string, prefix string) error {
	zapConf := zap.NewProductionConfig()
	switch strings.ToLower(lvl) {
	case "debug":
		l.lvl = zap.DebugLevel
		zapConf.Level = zap.NewAtomicLevelAt(l.lvl)
	case "info":
		l.lvl = zap.InfoLevel
		zapConf.Level = zap.NewAtomicLevelAt(l.lvl)
	case "error":
		l.lvl = zap.WarnLevel
		zapConf.Level = zap.NewAtomicLevelAt(l.lvl)
	case "critical":
		l.lvl = zap.ErrorLevel
		zapConf.Level = zap.NewAtomicLevelAt(l.lvl)
	}
	zapLog, err := zapConf.Build()
	if err != nil {
		return errors.Wrap(err)
	}
	l.SugaredLogger = zapLog.Sugar()
	l.prefix = prefix
	return nil
}

func Print(ctx context.Context, level int, v ...any) {
	if level < 1 || level > 4 || l.lvl > zapcore.Level(level-2) {
		return
	}
	requestId := requestid.Get(ctx)
	args := []any(nil)
	args = append(args, "app-name", l.prefix)
	if requestId != 0 {
		args = append(args, "request-id", requestId)
	}
	args = append(args, v...)

	switch level {
	case 1:
		l.Debug(args...)
	case 2:
		l.Info(args...)
	case 3:
		l.Warn(args...)
	case 4:
		l.Error(args...)
	}
}

func Debug(ctx context.Context, v ...any) {
	if l.lvl > zapcore.DebugLevel || l.SugaredLogger == nil {
		return
	}
	requestId := requestid.Get(ctx)
	args := []any(nil)
	args = append(args, "app-name", l.prefix)
	if requestId != 0 {
		args = append(args, "request-id", requestId)
	}
	args = append(args, v...)
	l.Debug(args...)
}

func Info(ctx context.Context, v ...any) {
	if l.lvl > zapcore.InfoLevel || l.SugaredLogger == nil {
		return
	}
	requestId := requestid.Get(ctx)
	args := []any(nil)
	args = append(args, "app-name", l.prefix)
	if requestId != 0 {
		args = append(args, "request-id", requestId)
	}
	args = append(args, v...)
	l.Info(args...)
}

func Error(ctx context.Context, v ...any) {
	if l.lvl > zapcore.WarnLevel || l.SugaredLogger == nil {
		return
	}
	requestId := requestid.Get(ctx)
	args := []any(nil)
	args = append(args, "app-name", l.prefix)
	if requestId != 0 {
		args = append(args, "request-id", requestId)
	}
	args = append(args, v...)
	l.Warn(args...)
}

func Critical(ctx context.Context, v ...any) {
	if l.lvl > zapcore.ErrorLevel || l.SugaredLogger == nil {
		return
	}
	requestId := requestid.Get(ctx)
	args := []any(nil)
	args = append(args, "app-name", l.prefix)
	if requestId != 0 {
		args = append(args, "request-id", requestId)
	}
	args = append(args, v...)
	l.Error(args...)
}
