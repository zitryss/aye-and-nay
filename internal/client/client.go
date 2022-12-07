package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func New(apiAddress string, timeout time.Duration, opts ...options) (*Client, error) {
	c := Client{}
	httpTransport := &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		ForceAttemptHTTP2: true,
		MaxIdleConns:      100,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	transport := newTransport(httpTransport, &c.m, &c.passed, &c.failed)
	httpClient := &http.Client{
		Jar:       nil,
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	c.client = httpClient
	c.apiAddress = apiAddress
	for _, opt := range opts {
		opt(&c)
	}
	if c.testdata != "" {
		err := c.readFiles()
		if err != nil {
			return &Client{}, errors.Wrap(err)
		}
	}
	return &c, nil
}

type options func(*Client)

func WithFiles(testdata string) options {
	return func(c *Client) {
		c.testdata = testdata
	}
}

func WithTimes(times int) options {
	return func(c *Client) {
		c.times = times
	}
}

type Client struct {
	client     *http.Client
	testdata   string
	apiAddress string
	times      int
	b          []byte
	sep        string
	m          sync.Mutex
	passed     int
	failed     int
}

func (c *Client) readFiles() error {
	body := bytes.Buffer{}
	multi := multipart.NewWriter(&body)
	for i := 0; i < c.times; i++ {
		for _, filename := range []string{"alan.jpg", "john.bmp", "dennis.png"} {
			part, err := multi.CreateFormFile("images", filename)
			if err != nil {
				return errors.Wrap(err)
			}
			b, err := os.ReadFile(c.testdata + "/" + filename)
			if err != nil {
				return errors.Wrap(err)
			}
			_, err = part.Write(b)
			if err != nil {
				return errors.Wrap(err)
			}
		}
	}
	err := multi.WriteField("duration", "1h")
	if err != nil {
		return errors.Wrap(err)
	}
	err = multi.Close()
	if err != nil {
		return errors.Wrap(err)
	}
	c.b = body.Bytes()
	c.sep = multi.FormDataContentType()
	return nil
}

func (c *Client) Album() (string, error) {
	body := bytes.NewReader(c.b)
	req, err := http.NewRequest(http.MethodPost, c.apiAddress+"/api/albums/", body)
	if err != nil {
		return "", errors.Wrap(err)
	}
	req.Header.Set("Content-Type", c.sep)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode/100 != 2 {
		return "", errors.Wrap(errors.New("response status code: expected = 2xx, actual = " + strconv.Itoa(resp.StatusCode)))
	}

	type result struct {
		Album struct {
			Id string
		}
	}

	res := result{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", errors.Wrap(err)
	}

	return res.Album.Id, nil
}

func (c *Client) Status(album string) error {
	req, err := http.NewRequest(http.MethodGet, c.apiAddress+"/api/albums/"+album+"/status/", http.NoBody)
	if err != nil {
		return errors.Wrap(err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode/100 != 2 {
		return errors.Wrap(errors.New("response status code: expected = 2xx, actual = " + strconv.Itoa(resp.StatusCode)))
	}

	type result struct {
		Album struct {
			Progress float64
		}
	}

	res := result{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

type Pair struct {
	One elem
	Two elem
}

type elem struct {
	Token string
	Src   string
}

func (c *Client) Pair(album string) (Pair, error) {
	req, err := http.NewRequest(http.MethodGet, c.apiAddress+"/api/albums/"+album+"/pair/", http.NoBody)
	if err != nil {
		return Pair{}, errors.Wrap(err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return Pair{}, errors.Wrap(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode/100 != 2 {
		return Pair{}, errors.Wrap(errors.New("response status code: expected = 2xx, actual = " + strconv.Itoa(resp.StatusCode)))
	}

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
	if err != nil {
		return Pair{}, errors.Wrap(err)
	}

	p := Pair{
		One: elem{
			Token: res.Album.Img1.Token,
			Src:   res.Album.Img1.Src,
		},
		Two: elem{
			Token: res.Album.Img2.Token,
			Src:   res.Album.Img2.Src,
		},
	}
	return p, nil
}

func (c *Client) Vote(album string, token1 string, token2 string) error {
	body := strings.NewReader("{\"album\":{\"imgFrom\":{\"token\":\"" + token1 + "\"},\"imgTo\":{\"token\":\"" + token2 + "\"}}}")
	req, err := http.NewRequest(http.MethodPatch, c.apiAddress+"/api/albums/"+album+"/vote/", body)
	if err != nil {
		return errors.Wrap(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode/100 != 2 {
		return errors.Wrap(errors.New("response status code: expected = 2xx, actual = " + strconv.Itoa(resp.StatusCode)))
	}

	return nil
}

func (c *Client) Top(album string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, c.apiAddress+"/api/albums/"+album+"/top/", http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode/100 != 2 {
		return nil, errors.Wrap(errors.New("response status code: expected = 2xx, actual = " + strconv.Itoa(resp.StatusCode)))
	}

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
	if err != nil {
		return nil, errors.Wrap(err)
	}

	src := []string(nil)
	for _, image := range res.Album.Images {
		src = append(src, image.Src)
	}
	return src, nil
}

func (c *Client) Health() error {
	req, err := http.NewRequest(http.MethodGet, c.apiAddress+"/api/health/", http.NoBody)
	if err != nil {
		return errors.Wrap(err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode/100 != 2 {
		return errors.Wrap(errors.New("response status code: expected = 2xx, actual = " + strconv.Itoa(resp.StatusCode)))
	}
	return nil
}

func (c *Client) Do(method string, url string, body io.ReadCloser) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.Wrap(err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode/100 != 2 {
		return errors.Wrap(errors.New("response status code: expected = 2xx, actual = " + strconv.Itoa(resp.StatusCode)))
	}
	return nil
}

func (c *Client) Stats() (int, int) {
	p := 0
	f := 0
	c.m.Lock()
	p = c.passed
	f = c.failed
	c.m.Unlock()
	return p, f
}
