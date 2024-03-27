package httprouter

import "net/http"

type RouterBuilder struct {
	prefix string
}

type SubRouterBuilder struct {
	parent Router
	prefix string
}

type RouteBuilder struct {
	middleware Middleware
	router     Router
}

func Prefix(p string) RouterBuilder {
	return RouterBuilder{
		prefix: p,
	}
}

func (rb RouterBuilder) NewRouter() *ServerRouter {
	return &ServerRouter{
		prefix:          rb.prefix,
		middlewares:     []Middleware{},
		methodNodes:     []*RouterNode{},
		notFoundHandler: notFound,
	}
}

func (rb SubRouterBuilder) SubRouter() Router {
	return &SubRouter{
		prefix:      rb.prefix,
		parent:      rb.parent,
		middlewares: []Middleware{},
	}
}

func (rb RouteBuilder) Get(route string, routeHandler http.HandlerFunc) {
	rb.router.Route("GET", route, routeHandler)
}

func (rb RouteBuilder) Post(route string, routeHandler http.HandlerFunc) {
	rb.router.Route("POST", route, routeHandler)
}

func (rb RouteBuilder) Put(route string, routeHandler http.HandlerFunc) {
	rb.router.Route("PUT", route, routeHandler)
}

func (rb RouteBuilder) Delete(route string, routeHandler http.HandlerFunc) {
	rb.router.Route("DELETE", route, routeHandler)
}

func (rb RouteBuilder) Route(method string, route string, routeHandler http.HandlerFunc) {
	routeHandler = buildHandler(routeHandler, 0, rb.middleware).ServeHTTP
	rb.router.Route(method, route, routeHandler)
}
