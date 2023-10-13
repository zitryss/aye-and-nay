package client

import (
	"net/http"
	"sync"
)

func newTransport(roundTrip http.RoundTripper, m *sync.Mutex, passed *int, failed *int) *transport {
	t := transport{}
	t.original = roundTrip
	t.m = m
	t.passed = passed
	t.failed = failed
	return &t
}

type transport struct {
	original http.RoundTripper
	m        *sync.Mutex
	passed   *int
	failed   *int
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.original.RoundTrip(req)
	if err != nil {
		t.m.Lock()
		*t.failed++
		t.m.Unlock()
		return resp, err
	}
	if resp.StatusCode/100 != 2 {
		t.m.Lock()
		*t.failed++
		t.m.Unlock()
		return resp, err
	}
	t.m.Lock()
	*t.passed++
	t.m.Unlock()
	return resp, err
}
