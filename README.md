# TigerMux
A simple router for building http services in Go! Supports path parameters, regex matching, subrouting, and middlewares! Includes utility functions to extract path parameters in handler functions. TigerMux is implemented using a Prefix tree and uses a recursive algorithm to construct routes with middleware.

This project was inspired by popular Go routers such as Chi and GorillaMux.
TigerMux was implemented as my first Go project mainly for the purposes of learning Go!

## Usage

We can start off by making a router like this,
```
router := mux.Prefix("/api").NewRouter()
```
The "Prefix" function specifies that all routes in the router will start with "/api"

We can add a middleware to our router with the "Use" function
```
router.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Before first middleware")
        next.ServeHTTP(w, r)
        log.Printf("After first middleware")
    })
})
```
In Go, a middleware is a function that takes a http.Handler as an argument (the next handler) and returns a http.Handler.
All requests to the router will pass the middleware. In the example above next.ServeHttp(w, r) calls the next handler in the chain which
could be another middleware or the handler.

We can define a route using the "Get", "Post", "Put", "Delete", or "Route" functions.
```
router.Get("/hello/world", func(w http.ResponseWriter, r *http.Request) {
    log.Printf("Get Hello World")
    w.Write([]byte("Hello World"))
})

router.Route("POST", "/hello/world", func(w http.ResponseWriter, r *http.Request) {
    log.Printf("Post Hello World")
    w.Write([]byte("Post Hello World"))
})
```

Routes can have path parameters which can be captured in the handler.
```
router.Get("/route/{param}/{param1}", func(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    log.Printf("Subrouter route param %s, %s", vars["param"], vars["param1"])
    w.Write([]byte("Subrouter route " + vars["param"] + " " + vars["param1"]))
})
```

We can specify routes that only match if a regex matches like so.
```
router.Get("/regex/|p([a-z]+)ch|", func(w http.ResponseWriter, r *http.Request) {
    log.Printf("Regex match")
    w.Write([]byte("Regex matched"))
})
```

If we want to apply a middleware directly to a route and that route only we can use the "With" function.
```
router.With(middleware).Get("/hello", func(w http.ResponseWriter, r *http.Request) {
    log.Printf("Subrouter Hello World")
    w.Write([]byte("Subrouter hello world"))
})
```

Lastly, we can create subroutes which contain their own middlewares and prefixes. All routes in a subrouter will inherit
the middleware and prefixes of their parent router. Subrouters can have their own subrouters.
```
subRouter := router.Prefix("/subroute").SubRouter()
subRouter.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Before subrouter middleware")
        next.ServeHTTP(w, r)
        log.Printf("After subrouter middleware")
    })
})
```

Examples can be seen in example.go at the root if you want an executable example.
