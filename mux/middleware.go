package mux

import (
	"log"
	"net/http"
)

func LoggerMiddleware(logger log.Logger) Middleware {
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