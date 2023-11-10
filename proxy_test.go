package gondola

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestServe(t *testing.T) {
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	backend1URL, err := url.Parse(backend1.URL)
	if err != nil {
		t.Fatal(err)
	}

	backend2URL, err := url.Parse(backend2.URL)
	if err != nil {
		t.Fatal(err)
	}

	data := `
proxy:
  port: 8080
  read_header_timeout: 20
  shutdown_timeout: 3000
upstreams:
  - host_name: backend1.local
    target: ` + backend1URL.String() + `
  - host_name: backend2.local
    target: ` + backend2URL.String() + `
`
	go func() {
		Run(strings.NewReader(data))
	}()

	// TODO: Find a better way to wait for the server to start.
	time.Sleep(time.Second)

	for _, test := range []struct {
		name    string
		reqPath string
		path    string
		body    string
		code    int
	}{
		{
			name:    "request to backend1",
			reqPath: "http://backend1.local/",
			body:    "backend1",
			code:    http.StatusOK,
		},
		{
			name:    "request to backend2",
			reqPath: "http://backend2.local/",
			body:    "backend2",
			code:    http.StatusOK,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, test.reqPath, nil)
			if err != nil {
				t.Fatal(err)
			}
			rec := httptest.NewRecorder()
			http.DefaultServeMux.ServeHTTP(rec, req)
			if rec.Body.String() != test.body {
				t.Errorf("Expected body %s, got %s", test.body, rec.Body.String())
			}
			if rec.Code != test.code {
				t.Errorf("Expected status code %d, got %d", test.code, rec.Code)
			}
		})
	}
}
