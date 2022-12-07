package compressor

import (
	"flag"
	"io"
	"os"
	"testing"

	"github.com/zitryss/aye-and-nay/internal/dockertest"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

var (
	unit        = flag.Bool("unit", false, "")
	integration = flag.Bool("int", false, "")
	ci          = flag.Bool("ci", false, "")
)

func TestMain(m *testing.M) {
	flag.Parse()
	if *ci || !*integration {
		code := m.Run()
		os.Exit(code)
	}
	log.SetOutput(os.Stderr)
	log.SetLevel(log.CRITICAL)
	docker := dockertest.New()
	host := &DefaultImaginaryConfig.Host
	port := &DefaultImaginaryConfig.Port
	docker.RunImaginary(host, port)
	log.SetOutput(io.Discard)
	code := m.Run()
	docker.Purge()
	os.Exit(code)
}
