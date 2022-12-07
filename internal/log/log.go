package log

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/internal/requestid"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

var (
	l logger
)

type logger struct {
	*zap.SugaredLogger
	lvl zapcore.Level
}

func New(lvl string, prefix string) error {
	zapConf := zap.NewProductionConfig()
	zapConf.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
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
	default:
		return errors.Wrap(errors.New("wrong log level"))
	}
	zapLog, err := zapConf.Build(zap.WithCaller(false))
	if err != nil {
		return errors.Wrap(err)
	}
	l.SugaredLogger = zapLog.Sugar()
	l.With("app-name", prefix)
	return nil
}

func newWithLogger(zapLog *zap.Logger, lvl zapcore.Level) {
	l.SugaredLogger = zapLog.Sugar()
	l.lvl = lvl
}

func Print(ctx context.Context, level int, msg string, v ...any) {
	switch level {
	case domain.LogDebug:
		Debug(ctx, msg, v...)
	case domain.LogInfo:
		Info(ctx, msg, v...)
	case domain.LogError:
		Error(ctx, msg, v...)
	case domain.LogCritical:
		Critical(ctx, msg, v...)
	}
}

func Debug(ctx context.Context, msg string, v ...any) {
	if l.lvl > zapcore.DebugLevel || l.SugaredLogger == nil {
		return
	}
	requestId := requestid.Get(ctx)
	if requestId != 0 {
		v = append([]any{"request-id", requestId}, v...)
	}
	l.Debugw(msg, v...)
}

func Info(ctx context.Context, msg string, v ...any) {
	if l.lvl > zapcore.InfoLevel || l.SugaredLogger == nil {
		return
	}
	requestId := requestid.Get(ctx)
	if requestId != 0 {
		v = append([]any{"request-id", requestId}, v...)
	}
	l.Infow(msg, v...)
}

func Error(ctx context.Context, msg string, v ...any) {
	if l.lvl > zapcore.WarnLevel || l.SugaredLogger == nil {
		return
	}
	requestId := requestid.Get(ctx)
	if requestId != 0 {
		v = append([]any{"request-id", requestId}, v...)
	}
	l.Warnw(msg, v...)
}

func Critical(ctx context.Context, msg string, v ...any) {
	if l.lvl > zapcore.ErrorLevel || l.SugaredLogger == nil {
		return
	}
	requestId := requestid.Get(ctx)
	if requestId != 0 {
		v = append([]any{"request-id", requestId}, v...)
	}
	l.Errorw(msg, v...)
}
