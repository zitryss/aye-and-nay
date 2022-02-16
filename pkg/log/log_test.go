package log_test

import (
	"flag"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zitryss/aye-and-nay/pkg/log"
)

var (
	unit        = flag.Bool("unit", false, "")
	integration = flag.Bool("int", false, "")
	ci          = flag.Bool("ci", false, "")
)

func TestLogLevelPositive(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	tests := []struct {
		level interface{}
		want  string
	}{
		{
			level: "debug",
			want:  "debug: message1\ndebug: message2: 60 95\ninfo: message3\ninfo: message4: mx 12\nerror: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n",
		},
		{
			level: "info",
			want:  "info: message3\ninfo: message4: mx 12\nerror: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n",
		},
		{
			level: "error",
			want:  "error: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n",
		},
		{
			level: "CRITICAL",
			want:  "critical: message7\ncritical: message8: ju iv\n",
		},
		{
			level: log.DEBUG,
			want:  "debug: message1\ndebug: message2: 60 95\ninfo: message3\ninfo: message4: mx 12\nerror: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n",
		},
		{
			level: log.INFO,
			want:  "info: message3\ninfo: message4: mx 12\nerror: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n",
		},
		{
			level: log.ERROR,
			want:  "error: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n",
		},
		{
			level: log.CRITICAL,
			want:  "critical: message7\ncritical: message8: ju iv\n",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			w := strings.Builder{}
			log.SetOutput(&w)
			log.SetPrefix("")
			log.SetFlags(0)
			log.SetLevel(tt.level)
			log.Debug("message1")
			log.Debugf("message2: %d %d", 60, 95)
			log.Info("message3")
			log.Infof("message4: %s %d", "mx", 12)
			log.Error("message5")
			log.Errorf("message6: %d %s", 80, "dq")
			log.Critical("message7")
			log.Criticalf("message8: %s %s", "ju", "iv")
			got := w.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLogLevelNegative(t *testing.T) {
	if !*unit {
		t.Skip()
	}
	tests := []struct {
		level interface{}
		want  string
	}{
		{
			level: 5,
			want:  "debug: message1\ndebug: message2: 60 95\ninfo: message3\ninfo: message4: mx 12\nerror: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n",
		},
		{
			level: "warning",
			want:  "debug: message1\ndebug: message2: 60 95\ninfo: message3\ninfo: message4: mx 12\nerror: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n",
		},
		{
			level: log.Level(-1),
			want:  "debug: message1\ndebug: message2: 60 95\ninfo: message3\ninfo: message4: mx 12\nerror: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			w := strings.Builder{}
			log.SetOutput(&w)
			log.SetPrefix("")
			log.SetFlags(0)
			log.SetLevel(log.DEBUG)
			log.Debug("message1")
			log.Debugf("message2: %d %d", 60, 95)
			log.Info("message3")
			log.Infof("message4: %s %d", "mx", 12)
			log.SetLevel(tt.level)
			log.Error("message5")
			log.Errorf("message6: %d %s", 80, "dq")
			log.Critical("message7")
			log.Criticalf("message8: %s %s", "ju", "iv")
			got := w.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
