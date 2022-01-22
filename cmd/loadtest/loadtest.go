package main

import (
	"net/http"
	"strings"

	"github.com/zitryss/aye-and-nay/internal/client"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

type loadtest struct {
	client *client.Client
	err    error
}

func (l *loadtest) albumApi() string {
	if l.err != nil {
		return ""
	}
	album, err := l.client.Album()
	if err != nil {
		l.err = errors.Wrap(err)
		log.Error(l.err)
		return ""
	}
	return album
}

func (l *loadtest) statusApi(album string) {
	if l.err != nil {
		return
	}
	err := l.client.Status(album)
	if err != nil {
		l.err = errors.Wrap(err)
		log.Error(l.err)
		return
	}
}

func (l *loadtest) pairApi(album string) client.Pair {
	if l.err != nil {
		return client.Pair{}
	}
	pairs, err := l.client.Pair(album)
	if err != nil {
		l.err = errors.Wrap(err)
		log.Error(l.err)
		return client.Pair{}
	}
	return pairs
}

func (l *loadtest) voteApi(album string, token1 string, token2 string) {
	if l.err != nil {
		return
	}
	err := l.client.Vote(album, token1, token2)
	if err != nil {
		l.err = errors.Wrap(err)
		log.Error(l.err)
		return
	}
}

func (l *loadtest) topApi(album string) []string {
	if l.err != nil {
		return nil
	}
	src, err := l.client.Top(album)
	if err != nil {
		l.err = errors.Wrap(err)
		log.Error(l.err)
		return nil
	}
	return src
}

func (l *loadtest) healthApi() {
	if l.err != nil {
		return
	}
	err := l.client.Health()
	if err != nil {
		l.err = errors.Wrap(err)
		log.Error(l.err)
		return
	}
}

func (l *loadtest) albumHtml() {
	l.html("/index.html")
}

func (l *loadtest) pairHtml() {
	l.html("/pair.html")
}

func (l *loadtest) topHtml() {
	l.html("/top.html")
}

func (l *loadtest) html(page string) {
	if l.err != nil {
		return
	}
	if htmlAddress == "" {
		return
	}
	err := l.client.Do(http.MethodGet, htmlAddress+page, http.NoBody)
	if err != nil {
		l.err = errors.Wrap(err)
		log.Error(l.err)
		return
	}
}

func (l *loadtest) pairMinio(src1 string, src2 string) {
	l.minio(src1)
	l.minio(src2)
}

func (l *loadtest) topMinio(src []string) {
	for _, s := range src {
		l.minio(s)
	}
}

func (l *loadtest) minio(src string) {
	if l.err != nil {
		return
	}
	address := minioAddress
	if strings.HasPrefix(src, "/api/images/") {
		address = apiAddress
	}
	if minioAddress == "" && address == "" {
		return
	}
	err := l.client.Do(http.MethodGet, address+src, http.NoBody)
	if err != nil {
		l.err = errors.Wrap(err)
		log.Error(l.err)
		return
	}
}
