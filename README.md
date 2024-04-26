# HttpRouter
A simple router for building http services in Go. Supports sub-routing and middlewares.

The package is implemented using a radix tree and uses a simple recursive algorithm to construct routes with middleware. It was inspired by popular Go routers such as [chi](https://github.com/go-chi/chi) and [gorilla/mux](https://github.com/gorilla/mux).
Package doesn't support path parameter matching - and instead uses a radix tree optimized for static paths.

## Install

With a [correctly configured](https://golang.org/doc/install#testing) Go toolchain:

```sh
go get -u github.com/JosephPrichard/HttpRouter
```


## Usage

Lets start off by registering a couple handlers to routes.
```go
r := httprouter.NewRouter()
r.Get("/", BaseHandler)
r.Post("/products", ProductsHandler)
r.Get("/articles", ArticlesHandler)
http.ListenAndServe(":9000", r)
```

We register three routes to 3 seperate handler functions. When the server listening on port `9000` receives a request - it will call the first handler function the url matches. The handler functions have the same function signature as `http.HandleFunc()`

We can also define a route using any request method as a string using the `Route` function.
```go
r.Route("POST", "/products", HandleCreateProduct)
r.Route("DELETE", "/products", HandleDeleteProduct)
```

### Prefixes

We can additionally add a prefix to our router.
```go
r := httprouter.Prefix("/api").NewRouter()
r.Get("/articles", ArticlesHandler)
http.ListenAndServe(":9000", r)
```

All routes attached to the router will only match a request url that includes the prefix.

### Middlewares

We can add a middleware to our router with the `Use` function.
```go
r.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Before this middleware")
        next.ServeHTTP(w, r)
        log.Printf("After this middleware")
    })
})
```

A middleware is a function that takes a `http.Handler` as an argument (the next handler) and returns a `http.Handler`.
All requests to the router will pass the middleware. In the example above `next.ServeHTTP(w, r)` calls the next handler in the chain which could be a handler or another middleware.

Middlewares will be added to any newly added routes they apply to. Adding a new middleware will not automatically add the middleware to any routes created beforehand, though.

If we want to apply a middleware directly to a route and that route only we can use the `With` function.
```go
r.With(middleware).Get("/products", HandleProduct)
r.With(middleware).Get("/articles", HandleArticle)
```

### Subrouters

Lastly, we can create subroutes which contain their own middlewares and prefixes.
```go
sr := r.Prefix("/subroute").SubRouter()
sr.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Before subrouter middleware")
        next.ServeHTTP(w, r)
        log.Printf("After subrouter middleware")
    })
})
```

All routes in a subrouter will inherit the middleware and prefixes of their parent router. Subrouters can have their own subrouters.

Note that subroutes will only inherit middlewares that exist when they are created. If we add a middleware to a parent route after we create a subrouter, the middleware will not be inherited automatically by the subroute.

### Runnable Example

```go 
package main

import (
    "net/http"
    "log"
    "github.com/JosephPrichard/HttpRouter"
)

func YourHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Pong."))
    w.WriteHeader(http.StatusOK)
}

func main() {
    r := httprouter.NewRouter()
    r.Get("/ping", YourHandler)
    log.Fatal(http.ListenAndServe(":8000", r))
}
```