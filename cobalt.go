// Package cobalt represents a small web toolkit to allow the building of web applications.
// It is primarily intended to be used for api web services. It allows the use of different encoders
// such as JSON, MsgPack, XML, etc..
//
// Context contains the http request and response writer. It also allows parameters to be added to the context as well. Context is passed to
// all prefilters, route handler and post filters. Context contains helper methods to extract the route parameters from the request.
package cobalt

import (
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/julienschmidt/httprouter"
)

type (
	// Coder is the interface used for the encoder in Cobalt. It allows the use
	// of multiple Encoders within cobalt
	Coder interface {
		Encode(w io.Writer, v interface{}) error
		Decode(r io.Reader, v interface{}) error
		ContentType() string
	}

	// Cobalt is the main data structure that holds all the filters, pointer to routes
	Cobalt struct {
		router      *httprouter.Router
		global      []MiddleWare
		serverError Handler
		coder       Coder
	}

	// Handler represents a request handler that is called by cobalt
	Handler func(c *Context)

	// MiddleWare is the type for middleware.
	MiddleWare func(Handler) Handler
)

// New creates a new instance of cobalt.
func New(coder Coder) *Cobalt {
	return &Cobalt{router: httprouter.New(), coder: coder}
}

// Coder returns the Coder configured in Cobalt
func (c *Cobalt) Coder() Coder {
	return c.coder
}

// ServerErr sets the handler for a server err.
func (c *Cobalt) ServerErr(h Handler) {
	c.serverError = h
}

// NotFound sets a not found handler.
func (c *Cobalt) NotFound(h Handler) {
	t := func(w http.ResponseWriter, req *http.Request) {
		ctx := NewContext(req, w, nil, c.coder)
		h(ctx)
	}

	c.router.NotFound = http.HandlerFunc(t)
}

// Route adds a route with an asscoiated method, handler and route filters.. It Builds a function which is then passed to the router.
func (c *Cobalt) route(method, route string, h Handler, m []MiddleWare) {

	f := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		st := time.Now()
		ctx := NewContext(req, w, p, c.coder)

		// Handle panics
		defer func() {
			if r := recover(); r != nil {
				log.Printf("cobalt: Panic, Recovering\n")
				log.Println(r)
				buf := make([]byte, 10000)
				runtime.Stack(buf, false)
				log.Printf("%s\n", string(buf))
				if c.serverError != nil {
					c.serverError(ctx)
				}
				if c.serverError == nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}

			log.Printf("Request %s complete [%s] =>  %s %s - %s", ctx.ID, time.Since(st), req.Method, req.RequestURI, req.RemoteAddr)
		}()

		log.Printf("Request %s start =>  %s %s - %s", ctx.ID, req.Method, req.RequestURI, req.RemoteAddr)

		w.Header().Set("X-Request-Id", ctx.ID)

		mwchain := func(h Handler) Handler {
			// global middleware.
			for idx := range c.global {
				h = c.global[idx](h)
			}

			// route specific middleware
			for idx := range m {
				h = m[idx](h)
			}
			return h
		}

		// process request
		mwchain(h)(ctx)
	}

	c.router.Handle(method, route, f)
}

// Get adds a route with an associated handler that matches a GET verb in a request.
func (c *Cobalt) Get(route string, h Handler, m ...MiddleWare) {
	c.route("GET", route, h, m)
}

// Post adds a route with an associated handler that matches a POST verb in a request.
func (c *Cobalt) Post(route string, h Handler, m ...MiddleWare) {
	c.route("POST", route, h, m)
}

// Put adds a route with an associated handler that matches a PUT verb in a request.
func (c *Cobalt) Put(route string, h Handler, m ...MiddleWare) {
	c.route("PUT", route, h, m)
}

// Delete adds a route with an associated handler that matches a DELETE verb in a request.
func (c *Cobalt) Delete(route string, h Handler, m ...MiddleWare) {
	c.route("DELETE", route, h, m)
}

// Options adds a route with an associated handler that matches a OPTIONS verb in a request.
func (c *Cobalt) Options(route string, h Handler, m ...MiddleWare) {
	c.route("OPTIONS", route, h, m)
}

// Head adds a route with an associated handler that matches a HEAD verb in a request.
func (c *Cobalt) Head(route string, h Handler, m ...MiddleWare) {
	c.route("HEAD", route, h, m)
}

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
func (c *Cobalt) ServeFiles(path string, root http.FileSystem) {
	c.router.ServeFiles(path, root)
}

// ServeHTTP implements the HandlerFunc that process the http request.
func (c *Cobalt) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c.router.ServeHTTP(w, req)
}

// Run runs the dispatcher which starts an http server to listen and serve.
func (c *Cobalt) Run(addr string) {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	log.SetPrefix("[cobalt] ")
	log.Printf("starting, listening on %s", addr)

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         addr,
		Handler:      c,
	}

	// TODO: add support for SSL/TLS
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf(err.Error())
	}
}

// RunTLS runs the dispatcher with a TLS cert.
func (c *Cobalt) RunTLS(addr, certfile, keyfile string) {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	log.SetPrefix("[cobalt] ")
	log.Printf("starting, listening on %s", addr)

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         addr,
		Handler:      c,
	}

	// TODO: add support for SSL/TLS
	if err := srv.ListenAndServeTLS(certfile, keyfile); err != nil {
		log.Fatalf(err.Error())
	}
}
