package log

import (
	"context"
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	unit        = flag.Bool("unit", true, "")
	integration = flag.Bool("int", false, "")
	ci          = flag.Bool("ci", false, "")
)

func testTimeEncoder(_ time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("")
}

func TestLogLevelPositive(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	tests := []struct {
		level zapcore.Level
		want  string
	}{
		{
			level: zapcore.DebugLevel,
			want:  "\tDEBUG\tmessage1\n\tINFO\tmessage3\n\tWARN\tmessage5\n\tERROR\tmessage7\n",
		},
		{
			level: zapcore.InfoLevel,
			want:  "\tINFO\tmessage3\n\tWARN\tmessage5\n\tERROR\tmessage7\n",
		},
		{
			level: zapcore.WarnLevel,
			want:  "\tWARN\tmessage5\n\tERROR\tmessage7\n",
		},
		{
			level: zapcore.ErrorLevel,
			want:  "\tERROR\tmessage7\n",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			encConf := zap.NewDevelopmentEncoderConfig()
			encConf.EncodeTime = testTimeEncoder
			enc := zapcore.NewConsoleEncoder(encConf)
			w := strings.Builder{}
			l := zap.New(zapcore.NewCore(enc, zapcore.AddSync(&w), tt.level))
			newWithLogger(l, tt.level)
			Debug(context.Background(), "message1")
			Info(context.Background(), "message3")
			Error(context.Background(), "message5")
			Critical(context.Background(), "message7")
			got := w.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
