// The purpose of this tool is to provide a realistic load on the
// system. It creates "n" albums (default 2), each contains 20 images.
// In total, it sends n x 94 requests.
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"

	"github.com/zitryss/aye-and-nay/pkg/debug"
)

var (
	n            int
	apiAddress   string
	minioAddress string
	htmlAddress  string
	connections  int
	testdata     string
	verbose      bool
	b            []byte
	sep          string
)

func main() {
	flag.IntVar(&n, "n", 2, "#albums")
	flag.StringVar(&apiAddress, "api-address", "https://localhost", "")
	flag.StringVar(&minioAddress, "minio-address", "https://localhost", "")
	flag.StringVar(&htmlAddress, "html-address", "https://localhost", "")
	flag.IntVar(&connections, "connections", 2, "")
	flag.StringVar(&testdata, "testdata", "./testdata", "")
	flag.BoolVar(&verbose, "verbose", true, "")
	flag.Parse()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	bar := pb.StartNew(n * 94)
	if !verbose {
		bar.SetWriter(io.Discard)
	}

	readFiles()

	sem := make(chan struct{}, connections)
	for i := 0; i < n; i++ {
		sem <- struct{}{}
		go func() {
			defer func() { <-sem }()
			albumHtml()
			album := albumApi()
			bar.Increment()
			readyApi(album)
			bar.Increment()
			for j := 0; j < 4; j++ {
				pairHtml()
				for k := 0; k < 11; k++ {
					src1, token1, src2, token2 := pairApi(album)
					bar.Increment()
					pairMinio(src1, src2)
					voteApi(album, token1, token2)
					bar.Increment()
				}
				topHtml()
				src := topApi(album)
				bar.Increment()
				topMinio(src)
			}
		}()
	}
	for i := 0; i < connections; i++ {
		sem <- struct{}{}
	}

	bar.Finish()
	fmt.Println(time.Since(bar.StartTime()))
	fmt.Println(float64(n*94)/time.Since(bar.StartTime()).Seconds(), "rps")
}

func readFiles() {
	body := bytes.Buffer{}
	multi := multipart.NewWriter(&body)
	for i := 0; i < 4; i++ {
		for _, filename := range []string{"alan.jpg", "john.bmp", "dennis.png", "tim.gif", "big.jpg"} {
			part, err := multi.CreateFormFile("images", filename)
			debug.Check(err)
			b, err := os.ReadFile(testdata + "/" + filename)
			debug.Check(err)
			_, err = part.Write(b)
			debug.Check(err)
		}
	}
	err := multi.WriteField("duration", "1h")
	debug.Check(err)
	err = multi.Close()
	debug.Check(err)

	b = body.Bytes()
	sep = multi.FormDataContentType()
}

func albumHtml() {
	html("/index.html")
}

func albumApi() string {
	body := bytes.NewReader(b)
	req, err := http.NewRequest(http.MethodPost, apiAddress+"/api/albums/", body)
	debug.Check(err)
	req.Header.Set("Content-Type", sep)

	resp, err := http.DefaultClient.Do(req)
	debug.Check(err)
	debug.Assert(resp.StatusCode == 201)

	type result struct {
		Album struct {
			Id string
		}
	}

	res := result{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	debug.Check(err)
	err = resp.Body.Close()
	debug.Check(err)

	return res.Album.Id
}

func readyApi(album string) {
	req, err := http.NewRequest(http.MethodGet, apiAddress+"/api/albums/"+album+"/ready/", http.NoBody)
	debug.Check(err)

	resp, err := http.DefaultClient.Do(req)
	debug.Check(err)
	debug.Assert(resp.StatusCode == 200)

	type result struct {
		Album struct {
			Progress float64
		}
	}

	res := result{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	debug.Check(err)
	err = resp.Body.Close()
	debug.Check(err)
}

func pairHtml() {
	html("/pair.html")
}

func pairApi(album string) (string, string, string, string) {
	req, err := http.NewRequest(http.MethodGet, apiAddress+"/api/albums/"+album+"/pair/", http.NoBody)
	debug.Check(err)

	resp, err := http.DefaultClient.Do(req)
	debug.Check(err)
	debug.Assert(resp.StatusCode == 200)

	type result struct {
		Album struct {
			Img1 struct {
				Token string
				Src   string
			}
			Img2 struct {
				Token string
				Src   string
			}
		}
	}

	res := result{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	debug.Check(err)
	err = resp.Body.Close()
	debug.Check(err)

	return res.Album.Img1.Src, res.Album.Img1.Token, res.Album.Img2.Src, res.Album.Img2.Token
}

func pairMinio(src1 string, src2 string) {
	minio(src1)
	minio(src2)
}

func voteApi(album string, token1 string, token2 string) {
	body := strings.NewReader("{\"album\":{\"imgFrom\":{\"token\":\"" + token1 + "\"},\"imgTo\":{\"token\":\"" + token2 + "\"}}}")
	req, err := http.NewRequest(http.MethodPatch, apiAddress+"/api/albums/"+album+"/vote/", body)
	debug.Check(err)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	debug.Check(err)
	debug.Assert(resp.StatusCode == 200)
	_, err = io.Copy(io.Discard, resp.Body)
	debug.Check(err)
	err = resp.Body.Close()
	debug.Check(err)
}

func topHtml() {
	html("/top.html")
}

func topApi(album string) []string {
	req, err := http.NewRequest(http.MethodGet, apiAddress+"/api/albums/"+album+"/top/", http.NoBody)
	debug.Check(err)

	resp, err := http.DefaultClient.Do(req)
	debug.Check(err)
	debug.Assert(resp.StatusCode == 200)

	type image struct {
		Src    string
		Rating float64
	}
	type result struct {
		Album struct {
			Images []image
		}
	}

	res := result{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	debug.Check(err)
	err = resp.Body.Close()
	debug.Check(err)

	src := []string(nil)
	for _, image := range res.Album.Images {
		src = append(src, image.Src)
	}
	return src
}

func topMinio(src []string) {
	for _, s := range src {
		minio(s)
	}
}

func html(page string) {
	if htmlAddress == "" {
		return
	}
	req, err := http.NewRequest(http.MethodGet, htmlAddress+page, http.NoBody)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	debug.Assert(resp.StatusCode == 200)
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

func minio(src string) {
	address := minioAddress
	if strings.HasPrefix(src, "/api/images/") {
		address = apiAddress
	}
	if minioAddress == "" && address == "" {
		return
	}
	req, err := http.NewRequest(http.MethodGet, address+src, http.NoBody)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	debug.Assert(resp.StatusCode == 200)
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}
