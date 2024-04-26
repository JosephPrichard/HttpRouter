package httprouter

import (
	"bytes"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func createTestRouter() *ServerRouter {
	r := Prefix("/api").NewRouter()
	sr := r.Prefix("/products").SubRouter()

	r.Use(LoggerMiddleware(log.Default()))
	r.Use(CorsMiddleware())

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("1 "))
			next.ServeHTTP(w, r)
			w.Write([]byte(" 1"))
		})
	})

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("2 "))

			b, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatalf("Failed to read bytes %s\n", err)
			}

			r.Body = io.NopCloser(bytes.NewReader(b))

			if string(b) == "Stop" {
				_, err := w.Write([]byte("Early Stop"))
				if err != nil {
					log.Fatalf("Failed to write message %s\n", err)
				}
			} else {
				next.ServeHTTP(w, r)
				w.Write([]byte(" 2"))
			}
		})
	})

	r.Post("/echo", func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatalf("Failed to read bytes %s\n", err)
		}
		_, err = w.Write(b)
		if err != nil {
			log.Fatalf("Failed to write message %s\n", err)
		}
	})

	sr.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Pong!"))
		if err != nil {
			log.Fatalf("Failed to write message %s\n", err)
		}
	})

	sr.Get("/ping/pong", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Ping Pong!"))
		if err != nil {
			log.Fatalf("Failed to write message %s\n", err)
		}
	})

	sr.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("3 "))
			next.ServeHTTP(w, r)
			w.Write([]byte(" 3"))
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Custom Not Found"))
		if err != nil {
			log.Fatalf("Failed to write message %s\n", err)
		}
	})

	return r
}

func TestRoutes(t *testing.T) {
	r := createTestRouter()

	expectedRoutes := []string{
		"POST /api/echo",
		"GET /api/products/ping",
		"GET /api/products/ping/pong",
	}
	actualRoutes := r.Routes()
	if !reflect.DeepEqual(actualRoutes, expectedRoutes) {
		t.Errorf("Expected matching routes but got %s", actualRoutes)
	}
}

func TestRouter(t *testing.T) {
	r := createTestRouter()

	server := httptest.NewServer(r)
	defer server.Close()

	type Test struct {
		method  string
		url     string
		bodyIn  string
		bodyOut string
	}

	testTable := []Test{
		{method: "GET", url: "/api/products/books/qbc", bodyIn: "", bodyOut: "Custom Not Found"},
		{method: "POST", url: "/api/echo", bodyIn: "Hello World", bodyOut: "1 2 Hello World 2 1"},
		{method: "POST", url: "/api/echo", bodyIn: "Stop", bodyOut: "1 2 Early Stop 1"},
		{method: "GET", url: "/api/products/ping", bodyIn: "", bodyOut: "1 2 Pong! 2 1"},
		{method: "GET", url: "/api/products/ping/pong", bodyIn: "", bodyOut: "1 2 Ping Pong! 2 1"},
	}

	fail := false
	for i, test := range testTable {
		func(test Test, i int) {
			var err error
			var resp *http.Response
			if test.method == "GET" {
				resp, err = http.Get(server.URL + test.url)
				if err != nil {
					t.Fatalf("Failed to send GET request: %v", err)
				}
			} else {
				reader := strings.NewReader(test.bodyIn)

				resp, err = http.Post(server.URL+test.url, "text/plain", reader)
				if err != nil {
					t.Fatalf("Failed to send POST request: %v", err)
				}
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			strBody := string(body)
			if strBody != test.bodyOut {
				t.Logf("Failed test %d, expected body %q, got %q", i, test.bodyOut, strBody)
				fail = true
			} else {
				t.Logf("Passed test %d", i)
			}
		}(test, i)
	}

	if fail {
		t.Fatalf("One or more route tests failed. Read logs.")
	}
}

func setupRoutes(b *testing.B) (*ServerRouter, []string) {
	b.ResetTimer()

	routes := generatePaths()
	r := NewRouter()

	for _, route := range routes {
		r.Get(route, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello World!"))
		})
	}

	b.StopTimer()
	b.Logf("Setting up routes took %d ms", b.Elapsed()/time.Millisecond)

	return r, routes
}

func BenchmarkSendRoutes(b *testing.B) {
	r, routes := setupRoutes(b)

	server := httptest.NewServer(r)
	defer server.Close()

	b.StartTimer()

	runCount := 1000
	fail := false

	for i := 0; i < runCount; i++ {
		func() {
			route := routes[rand.Intn(len(routes))]

			resp, err := http.Get(server.URL + route)
			if err != nil {
				b.Fatalf("Failed to send GET request: %v", err)
			}

			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				b.Fatalf("Failed to read response body: %v", err)
			}

			strBody := string(body)
			if strBody != "Hello World!" {
				b.Logf("Failed test %d, Expected Hello World!, got %s", i, strBody)
				fail = true
			} else {
				//b.Logf("Passed test %d", i)
			}
		}()
	}

	b.StopTimer()
	b.Logf("Sending requests took %d ms", b.Elapsed()/time.Millisecond)

	if fail {
		b.Fatalf("One or more route tests failed. Read logs.")
	}
}
