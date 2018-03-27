package cobalt

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
)

type (
	// Coder is the interface used for the encoder in Cobalt. It allows the use
	// of multiple Encoders within cobalt.
	Coder interface {
		Encode(w io.Writer, v interface{}) error
		Decode(r io.Reader, v interface{}) error
		ContentType() string
	}

	// Cobalt is the main data structure that holds all of the middleware and handlers.
	Cobalt struct {
		router      *httprouter.Router
		global      []MiddleWare
		serverError Handler
		cors        Handler
		coder       Coder

		// Templates is the configuration for HTML templates served by cobalt.
		Templates Templates
	}

	// Handler represents a request handler that is called by cobalt
	Handler func(c *Context)

	// MiddleWare is the type for middleware.
	MiddleWare func(Handler) Handler
)

// New creates a new instance of cobalt.
func New(coder Coder) *Cobalt {
	return &Cobalt{router: httprouter.New(), coder: coder, Templates: DefaultTemplates()}
}

// Coder returns the Coder configured in Cobalt
func (c *Cobalt) Coder() Coder {
	return c.coder
}

// CORS sets the handler for serving and processing cors.
func (c *Cobalt) CORS(h Handler) {
	c.cors = h
}

// ServerErr sets the handler for a server err.
func (c *Cobalt) ServerErr(h Handler) {
	c.serverError = h
}

// NotFound sets a not found handler.
func (c *Cobalt) NotFound(h Handler) {
	t := func(w http.ResponseWriter, req *http.Request) {
		ctx := NewContext(req, w, nil, c.coder, c.Templates)
		h(ctx)
	}

	c.router.NotFound = http.HandlerFunc(t)
}

// route adds a handler with middleware for a route and method. It builds a
// function which is then passed to the router.
func (c *Cobalt) route(method, route string, h Handler, m []MiddleWare) {

	f := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		st := time.Now()
		ctx := NewContext(req, w, p, c.coder, c.Templates)

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
// The path must end with "/*filepath", files are then served from the
// filesystem
//
// Example
//
//	c.ServeFiles("public/*filepath", http.Dir("public"))
func (c *Cobalt) ServeFiles(path string, root http.FileSystem) {
	c.router.ServeFiles(path, root)
}

// ServeHTTP implements http.Handler.
func (c *Cobalt) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// if method is options and handler set treat as preflight CORS request. Call the CORS handler.
	if c.cors != nil && req.Method == "OPTIONS" {
		ctx := NewContext(req, w, nil, c.coder, c.Templates)
		c.cors(ctx)
		return
	}

	// Otherwise just pass it on.
	c.router.ServeHTTP(w, req)
}

// Run runs the dispatcher which starts an http server to listen and serve.
// This operation blocks until an Inerrupt or SIGTERM signal is received at
// which point it starts a 10 second graceful shutdown. If requests do not
// finish in that time the whole application exits.
func (c *Cobalt) Run(addr string, readtimeout, writetimeout time.Duration) {
	c.run(addr, "", "", readtimeout, writetimeout)
}

// RunTLS runs the dispatcher with a TLS cert. It blocks waiting for a signal
// and performs graceful shutdown just like Run.
func (c *Cobalt) RunTLS(addr, certfile, keyfile string, readtimeout, writetimeout time.Duration) {
	c.run(addr, certfile, keyfile, readtimeout, writetimeout)
}

func (c *Cobalt) run(addr, certfile, keyfile string, readtimeout, writetimeout time.Duration) {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	log.SetPrefix("[cobalt] ")
	log.Printf("starting, listening on %s", addr)

	srv := &http.Server{
		ReadTimeout:  readtimeout,
		WriteTimeout: writetimeout,
		Addr:         addr,
		Handler:      c,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		if certfile == "" && keyfile == "" {
			serverErrors <- srv.ListenAndServe()
		} else {
			serverErrors <- srv.ListenAndServeTLS(certfile, keyfile)
		}
	}()

	shutdownTimeout := 10 * time.Second

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	// Blocking waiting for shutdown or an error
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case <-osSignals:

		// Create context for Shutdown call.
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown did not complete in %v : %v", shutdownTimeout, err)
			if err := srv.Close(); err != nil {
				log.Fatalf("Could not stop http server: %v", err)
			}
		}
	}
}
