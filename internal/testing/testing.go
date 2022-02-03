package testing

import (
	"bytes"
	_ "embed"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/base64"
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

func AlbumEmptyFactory(id uint64) model.Album {
	idB64 := base64.FromUint64(id)
	img1 := model.Image{Id: 0x3E3D, Src: "/aye-and-nay/albums/" + idB64 + "/images/PT4AAAAAAAA"}
	img2 := model.Image{Id: 0xB399, Src: "/aye-and-nay/albums/" + idB64 + "/images/mbMAAAAAAAA"}
	img3 := model.Image{Id: 0xDA2A, Src: "/aye-and-nay/albums/" + idB64 + "/images/KtoAAAAAAAA"}
	img4 := model.Image{Id: 0x51DE, Src: "/aye-and-nay/albums/" + idB64 + "/images/3lEAAAAAAAA"}
	img5 := model.Image{Id: 0xDA52, Src: "/aye-and-nay/albums/" + idB64 + "/images/UtoAAAAAAAA"}
	imgs := []model.Image{img1, img2, img3, img4, img5}
	edgs := map[uint64]map[uint64]int{}
	edgs[0x3E3D] = map[uint64]int{}
	edgs[0xB399] = map[uint64]int{}
	edgs[0xDA2A] = map[uint64]int{}
	edgs[0x51DE] = map[uint64]int{}
	edgs[0xDA52] = map[uint64]int{}
	expires := time.Time{}
	alb := model.Album{id, imgs, edgs, expires}
	return alb
}

func AlbumFullFactory(id uint64) model.Album {
	alb := AlbumEmptyFactory(id)
	alb.Images[0].Rating = 0.48954984
	alb.Images[1].Rating = 0.19186324
	alb.Images[2].Rating = 0.41218211
	alb.Images[3].Rating = 0.77920413
	alb.Images[4].Rating = 0.13278389
	alb.Edges[0x51DE][0xDA2A]++
	alb.Edges[0x3E3D][0xDA2A]++
	alb.Edges[0x3E3D][0x51DE]++
	alb.Edges[0xB399][0xDA2A]++
	alb.Edges[0xB399][0x51DE]++
	alb.Edges[0xB399][0x3E3D]++
	alb.Edges[0xDA52][0xDA2A]++
	alb.Edges[0xDA52][0x51DE]++
	alb.Edges[0xDA52][0x3E3D]++
	alb.Edges[0xDA52][0xB399]++
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
