package httprouter

import "net/http"

type Middleware = func(next http.Handler) http.Handler

type Router interface {
	Prefix(p string) SubRouterBuilder

	With(m Middleware) RouteBuilder

	SubRouter() Router

	Use(middleware Middleware)

	Route(method string, route string, routeHandler http.HandlerFunc)

	Get(route string, routeHandler http.HandlerFunc)

	Post(route string, routeHandler http.HandlerFunc)

	Put(route string, routeHandler http.HandlerFunc)

	Delete(route string, routeHandler http.HandlerFunc)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 not found"))
}

func buildHandler(baseHandler http.HandlerFunc, current int, middlewares ...Middleware) http.Handler {
	if current >= len(middlewares) {
		return baseHandler
	}
	middleware := middlewares[current]
	nextMiddleware := buildHandler(baseHandler, current+1, middlewares...)
	return middleware(nextMiddleware)
}
