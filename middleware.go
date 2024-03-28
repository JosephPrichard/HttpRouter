package httprouter

import (
	"log"
	"net/http"
)

type Middleware = func(next http.Handler) http.Handler

func LoggerMiddleware(logger *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Printf(
				"Header %s, Method %s, Path %s",
				r.Header,
				r.Method,
				r.URL.EscapedPath(),
			)
			next.ServeHTTP(w, r)
		})
	}
}

func CorsMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
