package http

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/strick-j/cybr-sdk-alpha/cybr"
)

func TestHTTPTransportBuilder_NoFollowRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Moved Permanently", http.StatusMovedPermanently)
	}))

	req, _ := http.NewRequest("GET", server.URL, nil)

	client := NewHTTPTransportBuilder()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if e, a := http.StatusMovedPermanently, resp.StatusCode; e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
}

func TestHTTPTransportBuilder_WithTimeout(t *testing.T) {
	client := &HTTPTransportBuilder{}

	expect := 10 * time.Millisecond
	client2 := client.WithTimeout(expect)

	if e, a := time.Duration(0), client.GetTimeout(); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}

	if e, a := expect, client2.GetTimeout(); e != a {
		t.Errorf("expected %v, got %v", e, a)
	}
}

func TestHTTPTransportBuild_concurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}))
	defer server.Close()

	var client cybr.HTTPClient = NewHTTPTransportBuilder()

	atOnce := 100
	var wg sync.WaitGroup
	wg.Add(atOnce)
	for i := 0; i < atOnce; i++ {
		go func(i int, client cybr.HTTPClient) {
			defer wg.Done()

			if v, ok := client.(interface{ GetTimeout() time.Duration }); ok {
				v.GetTimeout()
			}

			if i%3 == 0 {
				if v, ok := client.(interface {
					WithTransportOptions(opts ...func(*http.Transport)) cybr.HTTPClient
				}); ok {
					client = v.WithTransportOptions()
				}
			}

			req, _ := http.NewRequest("GET", server.URL, nil)
			resp, err := client.Do(req)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			resp.Body.Close()
		}(i, client)
	}
}
