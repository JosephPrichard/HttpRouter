package httprouter

import (
	"net/http"
	"strings"
)

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

type ServerRouter struct {
	prefix          string
	middlewares     []Middleware
	trie            trie[http.Handler]
	notFoundHandler http.HandlerFunc
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 not found"))
}

func buildHandler(baseHandler http.HandlerFunc, middlewares ...Middleware) http.Handler {
	if len(middlewares) == 0 {
		return baseHandler
	}
	middleware := middlewares[0]
	nextMiddleware := buildHandler(baseHandler, middlewares[1:]...)
	return middleware(nextMiddleware)
}

func NewRouter() *ServerRouter {
	return &ServerRouter{
		middlewares:     []Middleware{},
		trie:            newTrie[http.Handler](),
		notFoundHandler: notFound,
	}
}

func (router *ServerRouter) Use(middleware Middleware) {
	router.middlewares = append(router.middlewares, middleware)
}

func (router *ServerRouter) NotFound(routeHandler http.HandlerFunc) {
	router.notFoundHandler = routeHandler
}

func (router *ServerRouter) With(m Middleware) RouteBuilder {
	return RouteBuilder{
		middleware: m,
		router:     router,
	}
}

func (router *ServerRouter) Routes() []string {
	return router.trie.routes()
}

func (router *ServerRouter) Prefix(p string) SubRouterBuilder {
	return SubRouterBuilder{parent: router, prefix: p}
}

func (router *ServerRouter) SubRouter() Router {
	return &SubRouter{prefix: "", parent: router}
}

func (router *ServerRouter) Get(route string, routeHandler http.HandlerFunc) {
	router.Route("GET", route, routeHandler)
}

func (router *ServerRouter) Post(route string, routeHandler http.HandlerFunc) {
	router.Route("POST", route, routeHandler)
}

func (router *ServerRouter) Put(route string, routeHandler http.HandlerFunc) {
	router.Route("PUT", route, routeHandler)
}

func (router *ServerRouter) Delete(route string, routeHandler http.HandlerFunc) {
	router.Route("DELETE", route, routeHandler)
}

func (router *ServerRouter) Route(method string, route string, routeHandler http.HandlerFunc) {
	route = router.prefix + route
	handler := buildHandler(routeHandler, router.middlewares...)
	router.trie.insert(method, route, handler)
}

func (router *ServerRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.RequestURI, " \n\t")
	handler := router.trie.find(r.Method, path)
	if handler == nil || *handler == nil {
		router.notFoundHandler(w, r)
	} else {
		(*handler).ServeHTTP(w, r)
	}
}
