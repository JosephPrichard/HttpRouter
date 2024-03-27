package httprouter

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func CreateTestRouter() *ServerRouter {
	r := Prefix("/api").NewRouter()
	sr := r.Prefix("/products").SubRouter()

	r.Use(LoggerMiddleware(log.Default()))

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

	sr.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("3 "))
			next.ServeHTTP(w, r)
			w.Write([]byte(" 3"))
		})
	})

	sr.Get("/{name}/[[0-9]+]", func(w http.ResponseWriter, r *http.Request) {
		vars := Vars(r)
		name := vars["name"]
		_, err := w.Write([]byte("Name: " + name))
		if err != nil {
			log.Fatalf("Failed to write message %s\n", err)
		}
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
	r := CreateTestRouter()

	expectedRoutes := []string{
		"POST /api/echo",
		"GET /api/products/ping",
		"GET /api/products/{name}/[[0-9]+]",
	}
	actualRoutes := r.Routes()
	if !reflect.DeepEqual(actualRoutes, expectedRoutes) {
		t.Errorf("Expected 3 matching routes but got %s", actualRoutes)
	}
}

func TestRouter(t *testing.T) {
	r := CreateTestRouter()

	server := httptest.NewServer(r)
	defer server.Close()

	type Test struct {
		method  string
		url     string
		bodyIn  string
		bodyOut string
	}

	testTable := []Test{
		{method: "GET", url: "/api/products/books/123", bodyIn: "", bodyOut: "1 2 3 Name: books 3 2 1"},
		{method: "GET", url: "/api/products/books/qbc", bodyIn: "", bodyOut: "Custom Not Found"},
		{method: "POST", url: "/api/echo", bodyIn: "Hello World", bodyOut: "1 2 Hello World 2 1"},
		{method: "POST", url: "/api/echo", bodyIn: "Stop", bodyOut: "1 2 Early Stop 1"},
		{method: "GET", url: "/api/products/ping", bodyIn: "", bodyOut: "1 2 Pong! 2 1"},
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
