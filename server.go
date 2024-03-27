package httprouter

import (
	"log"
	"net/http"
	"strings"
)

type ServerRouter struct {
	prefix          string
	middlewares     []Middleware
	methodNodes     []*RouterNode
	notFoundHandler http.HandlerFunc
}

func NewRouter() *ServerRouter {
	return &ServerRouter{
		prefix:          "",
		middlewares:     []Middleware{},
		methodNodes:     []*RouterNode{},
		notFoundHandler: notFound,
	}
}

func (router *ServerRouter) With(m Middleware) RouteBuilder {
	return RouteBuilder{
		middleware: m,
		router:     router,
	}
}

func (router *ServerRouter) Routes() []string {
	var routes []string
	for _, m := range router.methodNodes {
		traverseNode(m.prefix+" ", m, &routes)
	}
	return routes
}

func (router *ServerRouter) Prefix(p string) SubRouterBuilder {
	return SubRouterBuilder{
		parent: router,
		prefix: p,
	}
}

func (router *ServerRouter) SubRouter() Router {
	return &SubRouter{
		prefix: "",
		parent: router,
	}
}

func (router *ServerRouter) findMethodNode(method string) *RouterNode {
	for _, node := range router.methodNodes {
		if node.prefix == method {
			return node
		}
	}
	node := &RouterNode{
		prefix:   method,
		children: []*RouterNode{},
	}
	router.methodNodes = append(router.methodNodes, node)
	return node
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
	node := router.findMethodNode(method)
	for _, prefix := range strings.Split(route, "/") {
		if prefix != "" {
			node = node.findOrCreateChild(prefix)
		}
	}
	node.handler = buildHandler(routeHandler, 0, router.middlewares...)
	log.Printf("Added route %s", route)
}

func (router *ServerRouter) Use(middleware Middleware) {
	router.middlewares = append(router.middlewares, middleware)
}

func (router *ServerRouter) NotFound(routeHandler http.HandlerFunc) {
	router.notFoundHandler = routeHandler
}

func (router *ServerRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	node := router.findMethodNode(r.Method)
	for _, prefix := range strings.Split(r.RequestURI, "/") {
		if prefix == "" {
			continue
		}
		node = node.matchChild(prefix)
		if node == nil {
			router.notFoundHandler(w, r)
			return
		}
		isParam, param := node.getURLParam()
		if isParam {
			setVar(r, param, prefix)
		}
	}
	if node.handler != nil {
		node.handler.ServeHTTP(w, r)
	} else {
		router.notFoundHandler(w, r)
	}
}
