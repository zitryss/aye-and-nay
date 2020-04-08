package log_test

import (
	"strings"
	"testing"

	"github.com/zitryss/aye-and-nay/pkg/log"
)

func TestLogLevelPositive(t *testing.T) {
	tests := []struct {
		level string
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
			if got != tt.want {
				t.Errorf("level = %v; got %v; want %v", tt.level, got, tt.want)
			}
		})
	}
	t.Run("", func(t *testing.T) {
		w := strings.Builder{}
		log.SetOutput(&w)
		log.SetPrefix("")
		log.SetFlags(0)
		log.SetLevel(log.Ldebug)
		log.Debug("message1")
		log.Debugf("message2: %d %d", 60, 95)
		log.Info("message3")
		log.Infof("message4: %s %d", "mx", 12)
		log.Error("message5")
		log.Errorf("message6: %d %s", 80, "dq")
		log.Critical("message7")
		log.Criticalf("message8: %s %s", "ju", "iv")
		got := w.String()
		want := "debug: message1\ndebug: message2: 60 95\ninfo: message3\ninfo: message4: mx 12\nerror: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n"
		if got != want {
			t.Errorf("level = %v; got %v; want %v", log.Ldebug, got, want)
		}
	})
	t.Run("", func(t *testing.T) {
		w := strings.Builder{}
		log.SetOutput(&w)
		log.SetPrefix("")
		log.SetFlags(0)
		log.SetLevel(log.Linfo)
		log.Debug("message1")
		log.Debugf("message2: %d %d", 60, 95)
		log.Info("message3")
		log.Infof("message4: %s %d", "mx", 12)
		log.Error("message5")
		log.Errorf("message6: %d %s", 80, "dq")
		log.Critical("message7")
		log.Criticalf("message8: %s %s", "ju", "iv")
		got := w.String()
		want := "info: message3\ninfo: message4: mx 12\nerror: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n"
		if got != want {
			t.Errorf("level = %v; got %v; want %v", log.Linfo, got, want)
		}
	})
	t.Run("", func(t *testing.T) {
		w := strings.Builder{}
		log.SetOutput(&w)
		log.SetPrefix("")
		log.SetFlags(0)
		log.SetLevel(log.Lerror)
		log.Debug("message1")
		log.Debugf("message2: %d %d", 60, 95)
		log.Info("message3")
		log.Infof("message4: %s %d", "mx", 12)
		log.Error("message5")
		log.Errorf("message6: %d %s", 80, "dq")
		log.Critical("message7")
		log.Criticalf("message8: %s %s", "ju", "iv")
		got := w.String()
		want := "error: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n"
		if got != want {
			t.Errorf("level = %v; got %v; want %v", log.Lerror, got, want)
		}
	})
	t.Run("", func(t *testing.T) {
		w := strings.Builder{}
		log.SetOutput(&w)
		log.SetPrefix("")
		log.SetFlags(0)
		log.SetLevel(log.Lcritical)
		log.Debug("message1")
		log.Debugf("message2: %d %d", 60, 95)
		log.Info("message3")
		log.Infof("message4: %s %d", "mx", 12)
		log.Error("message5")
		log.Errorf("message6: %d %s", 80, "dq")
		log.Critical("message7")
		log.Criticalf("message8: %s %s", "ju", "iv")
		got := w.String()
		want := "critical: message7\ncritical: message8: ju iv\n"
		if got != want {
			t.Errorf("level = %v; got %v; want %v", log.Lcritical, got, want)
		}
	})
}

func TestLogLevelNegative(t *testing.T) {
	t.Run("", func(t *testing.T) {
		w := strings.Builder{}
		log.SetOutput(&w)
		log.SetPrefix("")
		log.SetFlags(0)
		log.SetLevel(log.Ldebug)
		log.Debug("message1")
		log.Debugf("message2: %d %d", 60, 95)
		log.Info("message3")
		log.Infof("message4: %s %d", "mx", 12)
		log.SetLevel(5)
		log.Error("message5")
		log.Errorf("message6: %d %s", 80, "dq")
		log.Critical("message7")
		log.Criticalf("message8: %s %s", "ju", "iv")
		got := w.String()
		want := "debug: message1\ndebug: message2: 60 95\ninfo: message3\ninfo: message4: mx 12\nerror: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n"
		if got != want {
			t.Errorf("level = %v; got %v; want %v", log.Ldebug, got, want)
		}
	})
	t.Run("", func(t *testing.T) {
		w := strings.Builder{}
		log.SetOutput(&w)
		log.SetPrefix("")
		log.SetFlags(0)
		log.SetLevel(log.Ldebug)
		log.Debug("message1")
		log.Debugf("message2: %d %d", 60, 95)
		log.Info("message3")
		log.Infof("message4: %s %d", "mx", 12)
		log.SetLevel("warning")
		log.Error("message5")
		log.Errorf("message6: %d %s", 80, "dq")
		log.Critical("message7")
		log.Criticalf("message8: %s %s", "ju", "iv")
		got := w.String()
		want := "debug: message1\ndebug: message2: 60 95\ninfo: message3\ninfo: message4: mx 12\nerror: message5\nerror: message6: 80 dq\ncritical: message7\ncritical: message8: ju iv\n"
		if got != want {
			t.Errorf("level = %v; got %v; want %v", log.Ldebug, got, want)
		}
	})
}
