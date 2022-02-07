package testing

import (
	"bytes"
	_ "embed"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/zitryss/aye-and-nay/domain/model"
)

const (
	TOLERANCE = 0.000000000000001
)

var (
	//go:embed small.png
	png []byte
)

func Png() model.File {
	buf := bytes.NewBuffer(png)
	return model.File{Reader: buf, Size: int64(buf.Len())}
}

type ids interface {
	Uint64(i int) uint64
	Base64(i int) string
}

func AlbumEmptyFactory(id func() uint64, ids ids) model.Album {
	album := id()
	img1 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(1)}
	img2 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(2)}
	img3 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(3)}
	img4 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(4)}
	img5 := model.Image{Id: id(), Src: "/aye-and-nay/albums/" + ids.Base64(0) + "/images/" + ids.Base64(5)}
	imgs := []model.Image{img1, img2, img3, img4, img5}
	edgs := map[uint64]map[uint64]int{}
	edgs[ids.Uint64(1)] = map[uint64]int{}
	edgs[ids.Uint64(2)] = map[uint64]int{}
	edgs[ids.Uint64(3)] = map[uint64]int{}
	edgs[ids.Uint64(4)] = map[uint64]int{}
	edgs[ids.Uint64(5)] = map[uint64]int{}
	expires := time.Time{}
	alb := model.Album{album, imgs, edgs, expires}
	return alb
}

func AlbumFullFactory(id func() uint64, ids ids) model.Album {
	alb := AlbumEmptyFactory(id, ids)
	alb.Images[0].Rating = 0.48954984
	alb.Images[1].Rating = 0.19186324
	alb.Images[2].Rating = 0.41218211
	alb.Images[3].Rating = 0.77920413
	alb.Images[4].Rating = 0.13278389
	alb.Edges[ids.Uint64(4)][ids.Uint64(3)]++
	alb.Edges[ids.Uint64(1)][ids.Uint64(3)]++
	alb.Edges[ids.Uint64(1)][ids.Uint64(4)]++
	alb.Edges[ids.Uint64(2)][ids.Uint64(3)]++
	alb.Edges[ids.Uint64(2)][ids.Uint64(4)]++
	alb.Edges[ids.Uint64(2)][ids.Uint64(1)]++
	alb.Edges[ids.Uint64(5)][ids.Uint64(3)]++
	alb.Edges[ids.Uint64(5)][ids.Uint64(4)]++
	alb.Edges[ids.Uint64(5)][ids.Uint64(1)]++
	alb.Edges[ids.Uint64(5)][ids.Uint64(2)]++
	return alb
}

func AssertStatusCode(t *testing.T, w *httptest.ResponseRecorder, code int) {
	t.Helper()
	got := w.Code
	want := code
	if got != want {
		t.Errorf("Status Code = %v; want %v", got, want)
	}
}

func AssertContentType(t *testing.T, w *httptest.ResponseRecorder, content string) {
	t.Helper()
	got := w.Result().Header.Get("Content-Type")
	want := content
	if got != want {
		t.Errorf("Content-Type = %v; want %v", got, want)
	}
}

func AssertBody(t *testing.T, w *httptest.ResponseRecorder, body string) {
	t.Helper()
	got := w.Body.String()
	want := body
	if got != want {
		t.Errorf("Body = %v; want %v", got, want)
	}
}

func AssertChannel(t *testing.T, heartbeat <-chan interface{}) interface{} {
	t.Helper()
	v := interface{}(nil)
	select {
	case v = <-heartbeat:
	case <-time.After(1 * time.Second):
		t.Error("<-time.After(1 * time.Second)")
	}
	return v
}

func AssertNotChannel(t *testing.T, heartbeat <-chan interface{}) {
	t.Helper()
	select {
	case <-heartbeat:
		t.Error("<-heartbeatDel")
	case <-time.After(1 * time.Second):
	}
}

func AssertEqualFile(t *testing.T, x, y model.File) {
	t.Helper()
	bx := make([]byte, x.Size)
	_, err := x.Read(bx)
	assert.NoError(t, err)
	by := make([]byte, y.Size)
	_, err = y.Read(by)
	assert.NoError(t, err)
	assert.Equal(t, bx, by)
}
