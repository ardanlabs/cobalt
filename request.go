package cobalt

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/pborman/uuid"
)

const (
	// cacheControlHeader represents the http cache control header
	cacheControlHeader = "Cache-control"
)

type (

	// Request is the struct type that holds Request data for a request.
	// Request is scoped at the http request level, it is currently not Go routine safe for writes, so all writes
	// to Request should be done by 1 go routine
	Request struct {
		ID       string
		Response http.ResponseWriter
		Request  *http.Request
		Status   int
		// routevalues are the parameters in the route url from the http request.
		routevalues httprouter.Params
		coder       Coder
	}
)

var key = "request"

// FromContext retrieves a Request Context from context.Context
func FromContext(c context.Context) (*Request, bool) {
	request, ok := c.Value(key).(*Request)
	return request, ok
}

// ContextWith returns a new context with a request value.
func ContextWith(c context.Context, r *Request) context.Context {
	return context.WithValue(c, key, r)
}

// NewRequest creates a new Request instance with a http.Request and http.ResponseWriter.
func NewRequest(req *http.Request, resp http.ResponseWriter, p httprouter.Params, coder Coder) *Request {
	return &Request{
		ID:          uuid.New(),
		Request:     req,
		Response:    resp,
		routevalues: p,
		coder:       coder,
	}
}

// RouteValue returns the value for the associated key from the url parameters.
func (r *Request) RouteValue(key string) string {
	return r.routevalues.ByName(key)
}

// Error returns an http Error with the specified Error string and code
func (r *Request) Error(body interface{}, status int) {
	r.serveEncoded(body, 0, status)
}

// Decode decodes a reader into val
func (r *Request) Decode(reader io.Reader, val interface{}) error {
	return r.coder.Decode(reader, val)
}

// DecodeBody decodes a request body into val
func (r *Request) DecodeBody(val interface{}) error {
	return r.coder.Decode(r.Request.Body, val)
}

// Serve is a helper method to return encoded msg based on type from a struct type.
func (r *Request) Serve(val interface{}) {
	r.serveEncoded(val, http.StatusOK, 0)
}

// ServeWithStatus is a helper method to return encoded msg based on type from a struct type.
func (r *Request) ServeWithStatus(val interface{}, status int) {
	r.serveEncoded(val, status, 0)
}

// ServeStatus serves up the status passed in.
func (r *Request) ServeStatus(status int) {
	if status == 0 {
		status = http.StatusOK
	}
	r.Status = status
	r.Response.WriteHeader(r.Status)
}

// ServeCachedWithStatus is a helper method to return encoded msg based on type from a struct type.
func (r *Request) ServeCachedWithStatus(val interface{}, status int, seconds int) {
	r.serveEncoded(val, status, seconds)
}

// serveEncoded serves a value (val) encoded with expiring in seconds and a status
func (r *Request) serveEncoded(val interface{}, status int, seconds int) {
	if status == 0 {
		status = http.StatusOK
	}

	r.Response.Header().Set("Content-Type", r.coder.ContentType())
	if seconds > 0 {
		r.Response.Header().Set(cacheControlHeader, fmt.Sprintf("private, must-revalidate, max-age=%d", seconds))
	}

	r.Response.WriteHeader(status)

	if val != nil {
		if err := r.coder.Encode(r.Response, val); err != nil {
			r.Response.WriteHeader(http.StatusInternalServerError)
			status = http.StatusInternalServerError
		}
	}

	r.Status = status
}

// ServeResponse serves a response with the status and content type sent
func (r *Request) ServeResponse(resp []byte, status int, contentType string) {

	if contentType != "" {
		r.Response.Header().Set("Content-Type", contentType)
	}
	if contentType == "" {
		r.Response.Header().Set("Content-Type", r.coder.ContentType())
	}
	r.Response.WriteHeader(status)
	r.Response.Write(resp)
}
