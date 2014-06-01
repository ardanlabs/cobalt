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
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/OutCast-IO/pat"
)

const (
	GetMethod     = "GET"
	PostMethod    = "POST"
	PutMethod     = "PUT"
	DeleteMethod  = "DELETE"
	OptionsMethod = "OPTIONS"
	HeadMethod    = "HEAD"
)

var logger *log.Logger

type (
	Dispatcher struct {
		Router      *pat.PatternServeMux
		Prefilters  []FilterHandler
		Postfilters []Handler
	}

	Handler       func(c *Context)
	FilterHandler func(c *Context) bool
)

// init initializes the logger that used in the package.
func init() {
	logger = log.New(os.Stdout, "[cobalt] ", 0)
}

// NewDispatcher creates a new dispatcher.
func NewDispatcher() *Dispatcher {
	r := pat.New()
	return &Dispatcher{r, []FilterHandler{}, []Handler{}}
}

// AddPreFilter adds a prefilter hanlder to a dispatcher instance.
func (d *Dispatcher) AddPreFilter(h FilterHandler) {
	d.Prefilters = append(d.Prefilters, h)
}

// AddPostFilter adds a post processing handler to a diaptcher instance.
func (d *Dispatcher) AddPostFilter(h Handler) {
	d.Postfilters = append(d.Postfilters, h)
}

// Get adds a route with an associated handler that matches a GET verb in a request.
func (d *Dispatcher) Get(route string, h Handler) {
	d.addroute(GetMethod, route, h)
}

// Post adds a route with an associated handler that matches a POST verb in a request.
func (d *Dispatcher) Post(route string, h Handler) {
	d.addroute(PostMethod, route, h)
}

// Put adds a route with an associated handler that matches a PUT verb in a request.
func (d *Dispatcher) Put(route string, h Handler) {
	d.addroute(PutMethod, route, h)
}

// Delete adds a route with an associated handler that matches a DELETE verb in a request.
func (d *Dispatcher) Delete(route string, h Handler) {
	d.addroute(DeleteMethod, route, h)
}

// Options adds a route with an associated handler that matches a OPTIONS verb in a request.
func (d *Dispatcher) Options(route string, h Handler) {
	d.addroute(OptionsMethod, route, h)
}

// Head adds a route with an associated handler that matches a HEAD verb in a request.
func (d *Dispatcher) Head(route string, h Handler) {
	d.addroute(HeadMethod, route, h)
}

// addRoute adds a route with an asscoiated method and handler. It Builds a function which is then passed to the router.
func (d *Dispatcher) addroute(method, route string, h Handler) {

	f := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r := recover(); r != nil {
			buf := make([]byte, 10000)
			runtime.Stack(buf, false)
			logger.Printf("%s", string(buf))
		}

		ctx := NewContext(req, w)

		for _, filter := range d.Prefilters {
			exit := filter(ctx)
			if exit {
				return
			}
		}

		// call route handler
		h(ctx)

		for _, filter := range d.Postfilters {
			filter(ctx)
		}

	})

	d.Router.Add(method, route, f)
}

// Run runs the dispatcher which starts an http server to listen and serve.
func (d *Dispatcher) Run(addr string) {
	logger.Printf("starting, listening on %s", addr)

	http.Handle("/", d.Router)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		logger.Fatal(err)
	}
}
