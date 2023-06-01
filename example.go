package main

import (
	"Router/mux"
	"io"
	"log"
	"net/http"
)

func main() {
	log.Printf("Started router example")

	router := mux.Prefix("/api").NewRouter()
	subRouter := router.Prefix("/subroute").SubRouter()

	router.Use(mux.LoggerMiddleware(*log.Default()))

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Before first middleware")
			next.ServeHTTP(w, r)
			log.Printf("After first middleware")
		})
	})

	router.Get("/goodbye/world", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Get Goodbye World")
		w.Write([]byte("Goodbye World"))
	})

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Before second middlware")
			b, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatalf("Failed to read bytes %s\n", err)
			}
			if string(b) == "Stop" {
				log.Printf("Encountered stop message in second middleware")
			} else {
				next.ServeHTTP(w, r)
				log.Printf("After second middleware")
			}
		})
	})

	router.Get("/hello/world", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Get Hello World")
		w.Write([]byte("Hello World"))
	})

	router.Get("/hello/echo", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Echo")
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatalf("Failed to read bytes %s\n", err)
		}
		log.Printf(string(b))
		w.Write(b)
	})

	router.Post("/hello/world", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Post Hello World")
		w.Write([]byte("Created hello world"))
	})

	router.Get("/regex/|p([a-z]+)ch|", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Regex match")
		w.Write([]byte("Regex matched"))
	})

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Before third middleware")
			next.ServeHTTP(w, r)
			log.Printf("After third middleware")
		})
	})

	subRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Before fourth middleware")
			next.ServeHTTP(w, r)
			log.Printf("After fourth middleware")
		})
	})

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Before fifth middleware")
			next.ServeHTTP(w, r)
			log.Printf("After fifth middleware")
		})
	}

	subRouter.With(middleware).Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Subrouter Hello World")
		w.Write([]byte("Subrouter hello world"))
	})

	subRouter.Get("/route/{param}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		log.Printf("Subrouter route param %s", vars["param"])
		w.Write([]byte("Subrouter route " + vars["param"]))
	})

	subRouter.Get("/route/varr", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Varr match")
		w.Write([]byte("Varr"))
	})

	// test the route ordering
		router.Get("/test/{param}", func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			log.Printf("Testing param %s", vars["param"])
			w.Write([]byte("Testing param " + vars["param"]))
		})

		router.Get("/test/peach", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Testing literal 1")
			w.Write([]byte("Testing literal 1"))
		})

		router.Get("/test/|p([a-z]+)ch|", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Testing regex 1")
			w.Write([]byte("Testing regex 1"))
		})

		router.Get("/test/paech", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Testing literal 2")
			w.Write([]byte("Testing literal 2"))
		})

		router.Get("/test/|t([a-z]+)t|", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Testing regex 2")
			w.Write([]byte("Testing regex 2"))
		})
	//

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Custom Not Found"))
	})

	for _, r := range router.Routes() {
		log.Printf(r)
	}

	err := http.ListenAndServe(":9000", router)
	if err != nil {
		log.Fatalf("Couldn't start the server %s\n", err)
	}
}