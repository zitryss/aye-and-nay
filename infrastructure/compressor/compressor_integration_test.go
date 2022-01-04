//go:build integration

package compressor

import (
	"io"
	"os"
	"testing"

	"github.com/zitryss/aye-and-nay/internal/dockertest"
	"github.com/zitryss/aye-and-nay/pkg/env"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func TestMain(m *testing.M) {
	_, err := env.Lookup("CONTINUOUS_INTEGRATION")
	if err != nil {
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
	code := m.Run()
	os.Exit(code)
}
