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

func (router *ServerRouter) getOrCreateMethodNode(method string) *RouterNode {
	// iterate through level to find node with methods
	for _, node := range router.methodNodes {
		if node.prefix == method {
			return node
		}
	}
	// method node wasn't found, then create and return it
	node := &RouterNode{
		prefix:     method,
		childNodes: []*RouterNode{},
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
	// get the right method node for the route
	node := router.getOrCreateMethodNode(method)
	// climb down the tree based on the route prefixes
	for _, prefix := range strings.Split(route, "/") {
		if prefix != "" {
			node = node.findOrCreateChild(prefix)
		}
	}
	// add the handler to the final node with by creating handlers with middleware
	node.handler = buildHandler(routeHandler, 0, router.middlewares...)
	log.Printf("Added route %s to node %s", route, node.prefix)
}

func (router *ServerRouter) Use(middleware Middleware) {
	router.middlewares = append(router.middlewares, middleware)
}

func (router *ServerRouter) NotFound(routeHandler http.HandlerFunc) {
	router.notFoundHandler = routeHandler
}

func (router *ServerRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the right method node for the route
	node := router.getOrCreateMethodNode(r.Method)
	// climb down the tree based on the route prefixes, add url params along the way
	for _, prefix := range strings.Split(r.RequestURI, "/") {
		if prefix == "" {
			continue
		}
		node = node.matchChild(prefix)
		// node doesn't exist for this prefix so early return
		if node == nil {
			router.notFoundHandler(w, r)
			return
		}
		// check if is a url param, if so then add the param to request
		isParam, param := node.getURLParam()
		if isParam {
			setVar(r, param, prefix)
		}
	}
	// check if the node has a handler, and if so execute
	if node.handler != nil {
		log.Printf(node.prefix)
		node.handler.ServeHTTP(w, r)
	} else {
		router.notFoundHandler(w, r)
	}
}
