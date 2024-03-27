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
		router: router,
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
	routeHandler = buildHandler(routeHandler, 0, router.middlewares...).ServeHTTP
	router.parent.Route(method, route, routeHandler)
}

func (router *SubRouter) Use(middleware Middleware) {
	router.middlewares = append(router.middlewares, middleware)
}