// cobalt package represents a small web toolkit to allow the building of web applications.
// It is primarily intended to be used for JSON based API. It has routing, pre-filters and post filters.
//
// Pre-filters are called after the router identifies the proper route and before the user code (handler) is called.
// Pre-filters allow you to write to the response and end the request chain by returning a value of true from the filter handler.
//
// Post filters allow you to specify a handler that gets called after the user code (handler) is run.
//
// Context contains the http request and response writer. It also allows parameters to be added to the context as well. Context is passed to
// all prefilters, route handler and post filters. Context contains helper methods to extract the route parameters from the request.
package cobalt

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/OutCast-IO/httptreemux"
)

const (
	GetMethod     = "GET"
	PostMethod    = "POST"
	PutMethod     = "PUT"
	DeleteMethod  = "DELETE"
	OptionsMethod = "OPTIONS"
	HeadMethod    = "HEAD"
)

type (
	Cobalt struct {
		router          *httptreemux.TreeMux
		prefilters      []FilterHandler
		postfilters     []Handler
		notFoundHandler Handler
		serverError     Handler
	}

	Handler       func(c *Context)
	FilterHandler func(c *Context) bool
)

// NewDispatcher creates a new dispatcher.
func New() *Cobalt {
	r := httptreemux.New()
	return &Cobalt{r, []FilterHandler{}, []Handler{}, nil, nil}
}

// AddPreFilter adds a prefilter hanlder to a dispatcher instance.
func (c *Cobalt) AddPrefilter(h FilterHandler) {
	c.prefilters = append(c.prefilters, h)
}

// AddPostFilter adds a post processing handler to a diaptcher instance.
func (c *Cobalt) AddPostfilter(h Handler) {
	c.postfilters = append(c.postfilters, h)
}

// Get adds a route with an associated handler that matches a GET verb in a request.
func (c *Cobalt) Get(route string, h Handler) {
	c.addroute(GetMethod, route, h)
}

// Post adds a route with an associated handler that matches a POST verb in a request.
func (c *Cobalt) Post(route string, h Handler) {
	c.addroute(PostMethod, route, h)
}

// Put adds a route with an associated handler that matches a PUT verb in a request.
func (c *Cobalt) Put(route string, h Handler) {
	c.addroute(PutMethod, route, h)
}

// Delete adds a route with an associated handler that matches a DELETE verb in a request.
func (c *Cobalt) Delete(route string, h Handler) {
	c.addroute(DeleteMethod, route, h)
}

// Options adds a route with an associated handler that matches a OPTIONS verb in a request.
func (c *Cobalt) Options(route string, h Handler) {
	c.addroute(OptionsMethod, route, h)
}

// Head adds a route with an associated handler that matches a HEAD verb in a request.
func (c *Cobalt) Head(route string, h Handler) {
	c.addroute(HeadMethod, route, h)
}

// ServeHTTP implements the HandlerFunc that process the http request.
func (c *Cobalt) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// future use for middleware.
	c.router.ServeHTTP(w, req)
}

// Run runs the dispatcher which starts an http server to listen and serve.
func (c *Cobalt) Run(addr string) {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	log.SetPrefix("[cobalt] ")
	log.Printf("starting, listening on %s", addr)

	//http.Handle("/", c.Router)
	err := http.ListenAndServe(addr, c)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

// addRoute adds a route with an asscoiated method and handler. It Builds a function which is then passed to the router.
func (c *Cobalt) addroute(method, route string, h Handler) {

	f := func(w http.ResponseWriter, req *http.Request, p map[string]string) {
		if r := recover(); r != nil {
			buf := make([]byte, 10000)
			runtime.Stack(buf, false)
			fmt.Printf("%s", string(buf))
		}

		ctx := NewContext(req, w, p)

		for i := 0; i < len(c.prefilters); i++ {
			exit := c.prefilters[i](ctx)
			if exit {
				return
			}
		}

		/*
			for _, filter := range c.Prefilters {
				exit := filter(ctx)
				if exit {
					return
				}
			}
		*/
		// call route handler
		h(ctx)

		for _, filter := range c.postfilters {
			filter(ctx)
		}
	}

	c.router.Handle(method, route, f)
}
