package httprouter

import "net/http"

type SubRouter struct {
	prefix      string
	middlewares []Middleware
	parent      Router
}

func (router *SubRouter) SubRouter() Router {
	return &SubRouter{
		prefix: "",
		parent: router,
	}
}

func (router *SubRouter) Prefix(p string) SubRouterBuilder {
	return SubRouterBuilder{
		parent: router,
		prefix: p,
	}
}

func (router *SubRouter) With(m Middleware) RouteBuilder {
	return RouteBuilder{
		middleware: m,
		router:     router,
	}
}

func (router *SubRouter) Get(route string, routeHandler http.HandlerFunc) {
	router.Route("GET", route, routeHandler)
}

func (router *SubRouter) Post(route string, routeHandler http.HandlerFunc) {
	router.Route("POST", route, routeHandler)
}

func (router *SubRouter) Put(route string, routeHandler http.HandlerFunc) {
	router.Route("PUT", route, routeHandler)
}

func (router *SubRouter) Delete(route string, routeHandler http.HandlerFunc) {
	router.Route("DELETE", route, routeHandler)
}

func (router *SubRouter) Route(method string, route string, routeHandler http.HandlerFunc) {
	route = router.prefix + route
	routeHandler = buildHandler(routeHandler, router.middlewares...).ServeHTTP
	router.parent.Route(method, route, routeHandler)
}

func (router *SubRouter) Use(middleware Middleware) {
	router.middlewares = append(router.middlewares, middleware)
}

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
		tree:            tree{},
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
	routeHandler = buildHandler(routeHandler, rb.middleware).ServeHTTP
	rb.router.Route(method, route, routeHandler)
}
