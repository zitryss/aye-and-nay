package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/zitryss/aye-and-nay/internal/client"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

var (
	duration     time.Duration
	connections  int
	timeout      time.Duration
	testdata     string
	apiAddress   string
	htmlAddress  string
	minioAddress string
	verbose      bool
)

func main() {
	flag.DurationVar(&duration, "duration", 10*time.Second, "in seconds")
	flag.IntVar(&connections, "connections", 2, "")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "in seconds")
	flag.StringVar(&testdata, "testdata", "./testdata", "")
	flag.StringVar(&apiAddress, "api-address", "https://localhost", "")
	flag.StringVar(&htmlAddress, "html-address", "https://localhost", "")
	flag.StringVar(&minioAddress, "minio-address", "https://localhost", "")
	flag.BoolVar(&verbose, "verbose", true, "")
	flag.Parse()

	if verbose {
		log.SetOutput(os.Stderr)
		log.SetLevel(log.ERROR)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	wg := sync.WaitGroup{}
	wg.Add(2)

	c, err := client.New(apiAddress, timeout*time.Second, client.WithFiles(testdata), client.WithTimes(5))
	if err != nil {
		log.Critical(err)
		os.Exit(1)
	}

	go func() {
		defer wg.Done()
		passed1sAgo := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
				passed, failed := c.Stats()
				fmt.Printf("%d rps, ", passed-passed1sAgo)
				fmt.Printf("%d passed (%.2f%%), ", passed, float64(passed)/float64(passed+failed)*100)
				fmt.Printf("%d failed (%.2f%%)\n", failed, float64(failed)/float64(passed+failed)*100)
				passed1sAgo = passed
			}
		}
	}()

	go func() {
		defer wg.Done()
		start := time.Now()
		sem := make(chan struct{}, connections)
		for time.Since(start) < duration*time.Second {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			go func() {
				defer func() { <-sem }()
				l := loadtest{client: c}
				l.albumHtml()
				album := l.albumApi()
				l.statusApi(album)
				for j := 0; j < 5; j++ {
					l.pairHtml()
					for k := 0; k < 10; k++ {
						pairs := l.pairApi(album)
						l.pairMinio(pairs.One.Src, pairs.Two.Src)
						l.voteApi(album, pairs.One.Token, pairs.Two.Token)
					}
					l.topHtml()
					src := l.topApi(album)
					l.topMinio(src)
				}
				l.healthApi()
			}()
		}
		for i := 0; i < connections; i++ {
			sem <- struct{}{}
		}
		stop()
	}()

	wg.Wait()

	_, failed := c.Stats()
	if failed > 0 {
		os.Exit(1)
	}
}
