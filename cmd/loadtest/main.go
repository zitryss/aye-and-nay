// The purpose of this tool is to provide a realistic load on the
// system. It creates 98 albums, each contains 20 images. In total, it
// sends 9996 requests.
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/zitryss/aye-and-nay/pkg/debug"
)

var (
	address     string
	connections int
	testdata    string
)

func main() {
	flag.StringVar(&address, "address", "https://localhost:8001", "")
	flag.IntVar(&connections, "connections", 25, "")
	flag.StringVar(&testdata, "testdata", "./testdata", "")
	flag.Parse()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	sem := make(chan struct{}, connections)
	for i := 0; i < 98; i++ {
		sem <- struct{}{}
		go func() {
			defer func() { <-sem }()
			session := albumHtml()
			id := albumApi(session)
			for j := 0; j < 4; j++ {
				pairHtml(session, id)
				for k := 0; k < 11; k++ {
					token1, token2 := pairApi(session, id)
					voteApi(session, id, token1, token2)
				}
				topHtml(session, id)
				topApi(session, id)
			}
		}()
	}
	for i := 0; i < connections; i++ {
		sem <- struct{}{}
	}
}

func albumHtml() string {
	resp, err := http.DefaultClient.Get(address + "/")
	debug.Check(err)
	debug.Assert(resp.StatusCode == 200)
	_, err = io.Copy(ioutil.Discard, resp.Body)
	debug.Check(err)
	err = resp.Body.Close()
	debug.Check(err)
	head := resp.Header.Get("Set-Cookie")
	str := strings.TrimPrefix(head, "session=")
	substr := strings.Split(str, ";")
	session := substr[0]
	return session
}

func albumApi(session string) string {
	body := bytes.Buffer{}
	multi := multipart.NewWriter(&body)
	for i := 0; i < 4; i++ {
		for _, filename := range []string{"alan.jpg", "john.bmp", "dennis.png", "tim.gif", "linus.jpg"} {
			part, err := multi.CreateFormFile("images", filename)
			debug.Check(err)
			b, err := ioutil.ReadFile(testdata + "/" + filename)
			debug.Check(err)
			_, err = part.Write(b)
			debug.Check(err)
		}
	}
	err := multi.Close()
	debug.Check(err)

	req, err := http.NewRequest("POST", address+"/api/albums/", &body)
	debug.Check(err)
	req.Header.Set("Content-Type", multi.FormDataContentType())
	c := http.Cookie{}
	c.Name = "session"
	c.Value = session
	req.AddCookie(&c)

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

func pairHtml(session string, id string) {
	req, err := http.NewRequest("GET", address+"/albums/"+id+"/", nil)
	debug.Check(err)
	c := http.Cookie{}
	c.Name = "session"
	c.Value = session
	req.AddCookie(&c)

	resp, err := http.DefaultClient.Do(req)
	debug.Check(err)
	debug.Assert(resp.StatusCode == 200)
	_, err = io.Copy(ioutil.Discard, resp.Body)
	debug.Check(err)
	err = resp.Body.Close()
	debug.Check(err)
}

func pairApi(session string, id string) (string, string) {
	req, err := http.NewRequest("GET", address+"/api/albums/"+id+"/", nil)
	debug.Check(err)
	c := http.Cookie{}
	c.Name = "session"
	c.Value = session
	req.AddCookie(&c)

	resp, err := http.DefaultClient.Do(req)
	debug.Check(err)
	debug.Assert(resp.StatusCode == 200)

	type result struct {
		Img1 struct {
			Token string
			Src   string
		}
		Img2 struct {
			Token string
			Src   string
		}
	}

	res := result{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	debug.Check(err)
	err = resp.Body.Close()
	debug.Check(err)

	return res.Img1.Token, res.Img2.Token
}

func voteApi(session string, id string, token1 string, token2 string) {
	body := strings.NewReader("{\"album\":{\"imgFrom\":{\"token\":\"" + token1 + "\"},\"imgTo\":{\"token\":\"" + token2 + "\"}}}")
	req, err := http.NewRequest("PATCH", address+"/api/albums/"+id+"/", body)
	debug.Check(err)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	c := http.Cookie{}
	c.Name = "session"
	c.Value = session
	req.AddCookie(&c)

	resp, err := http.DefaultClient.Do(req)
	debug.Check(err)
	debug.Assert(resp.StatusCode == 200)
	_, err = io.Copy(ioutil.Discard, resp.Body)
	debug.Check(err)
	err = resp.Body.Close()
	debug.Check(err)
}

func topHtml(session string, id string) {
	req, err := http.NewRequest("GET", address+"/albums/"+id+"/top/", nil)
	debug.Check(err)
	c := http.Cookie{}
	c.Name = "session"
	c.Value = session
	req.AddCookie(&c)

	resp, err := http.DefaultClient.Do(req)
	debug.Check(err)
	debug.Assert(resp.StatusCode == 200)
	_, err = io.Copy(ioutil.Discard, resp.Body)
	debug.Check(err)
	err = resp.Body.Close()
	debug.Check(err)
}

func topApi(session string, id string) {
	req, err := http.NewRequest("GET", address+"/api/albums/"+id+"/top/", nil)
	debug.Check(err)
	c := http.Cookie{}
	c.Name = "session"
	c.Value = session
	req.AddCookie(&c)

	resp, err := http.DefaultClient.Do(req)
	debug.Check(err)
	debug.Assert(resp.StatusCode == 200)
	_, err = io.Copy(ioutil.Discard, resp.Body)
	debug.Check(err)
	err = resp.Body.Close()
	debug.Check(err)
}
