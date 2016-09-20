// Package cobalt represents a small web toolkit to allow the building of web applications.
// It is primarily intended to be used for api web services. It allows the use of different encoders
// such as JSON, MsgPack, XML, etc..
//
// Response contains the http request and response writer. Request contains helper methods to extract
// the route parameters from the request and serve responses.
package cobalt

import (
	"context"
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
		serverError Handler
		coder       Coder
		// request timeout in milliseconds.
		timeout int
	}

	// Handler represents a request handler that is called by cobalt
	Handler func(r *Request)

	// MiddleWare is the type for middleware.
	MiddleWare func(Handler) Handler
)

const timeout = 10000 //default timeout set to 10 seconds.

// New creates a new instance of cobalt.
func New(coder Coder) *Cobalt {
	return &Cobalt{router: httprouter.New(), coder: coder, timeout: 10000}
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
		r := NewRequest(req, w, nil, c.coder)
		h(r)
	}

	c.router.NotFound = http.HandlerFunc(t)
}

// Route adds a route with an asscoiated method, handler and route filters.. It Builds a function which is then passed to the router.
func (c *Cobalt) route(method, route string, h Handler, m []MiddleWare) {

	f := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		st := time.Now()

		ctx, cancel := context.WithTimeout(req.Context(), 10*time.Second)
		req = req.WithContext(ctx)
		request := NewRequest(req, w, p, c.coder)

		// Handle panics, if server error handler is specified it will be called.
		// Otherwise, a generic 500 error will be sent. While the http package will
		// capture the panic, we capture it so we can serve the specified 500 error
		// that is configured for cobalt.
		defer func() {
			cancel()
			if r := recover(); r != nil {
				log.Printf("cobalt: Panic, Recovering\n")
				log.Println(r)
				buf := make([]byte, 10000)
				runtime.Stack(buf, false)
				log.Printf("%s\n", string(buf))
				request.Status = 500

				if c.serverError != nil {
					c.serverError(request)
				}
				if c.serverError == nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}

			log.Printf("Request %s complete [%s] status [%d] =>  %s %s - %s", request.ID, time.Since(st), request.Status, req.Method, req.RequestURI, req.RemoteAddr)
		}()

		log.Printf("Request %s start =>  %s %s - %s", request.ID, method, req.RequestURI, req.RemoteAddr)

		w.Header().Set("X-Request-Id", request.ID)

		mwchain := func(h Handler) Handler {
			// route middleware
			for idx := range m {
				h = m[idx](h)
			}
			return h
		}

		mwchain(h)(request)
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

	// TODO: add support for SSL/TLS
	err := http.ListenAndServe(addr, c)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
