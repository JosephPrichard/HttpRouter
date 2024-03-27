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
	w.Write([]byte("Route Not Found"))
}

// appends a slice of middlewares to the base handler function
func buildHandler(baseHandler http.HandlerFunc, current int, middlewares ...Middleware) http.Handler {
	// at last middleware so the next function is to serve base case
	if current >= len(middlewares) {
		return baseHandler
	}
	// recursively get the next middleware as the argument for the current middleware
	middleware := middlewares[current]
	nextMiddleware := buildHandler(baseHandler, current+1, middlewares...)
	return middleware(nextMiddleware)
}
