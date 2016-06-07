package cobalt

import (
	"fmt"
	"io"
	"net/http"

	"github.com/ardanlabs/cobalt/httprouter"
	"github.com/ardanlabs/cobalt/uuid"
)

const (
	// cacheControlHeader represents the http cache control header
	cacheControlHeader = "Cache-control"
)

type (

	// Context is the struct type that holds context data for a request.
	// Context is scoped at request level, it is currently not Go routine safe for writes, so all writes
	// to context should be done by 1 go routine
	Context struct {
		ID       string
		Response http.ResponseWriter
		Request  *http.Request
		Status   int
		// data that can be stored in the context for life of request
		data map[string]interface{}
		// params are the request parameters from the http request
		params httprouter.Params
		coder  Coder
	}
)

// NewContext creates a new context instance with a http.Request and http.ResponseWriter.
func NewContext(req *http.Request, resp http.ResponseWriter, p httprouter.Params, coder Coder) *Context {
	id, _ := uuid.NewV4()

	return &Context{
		ID:       id.String(),
		Request:  req,
		Response: resp,
		data:     make(map[string]interface{}),
		params:   p,
		coder:    coder,
	}
}

// ParamValue returns the value for the associated key from the url parameters.
func (c *Context) ParamValue(key string) string {
	return c.params.ByName(key)
}

// GetData returns the value for the specified key from the context data. Usually used by prefilters to pass data to the http handler
// and post filters.
func (c *Context) GetData(key string) interface{} {
	data, ok := c.data[key]
	if !ok {
		return nil
	}
	return data
}

// SetData sets the data for the specified key in the context instance.
func (c *Context) SetData(key string, value interface{}) {
	c.data[key] = value
}

// Error returns an http Error with the specified Error string and code
func (c *Context) Error(body interface{}, status int) {
	c.serveEncoded(body, 0, status)
}

// Decode decodes a reader into val
func (c *Context) Decode(r io.Reader, val interface{}) error {
	return c.coder.Decode(r, val)
}

// DecodeBody decodes a request body into val
func (c *Context) DecodeBody(val interface{}) error {
	return c.coder.Decode(c.Request.Body, val)
}

// Serve is a helper method to return encoded msg based on type from a struct type.
func (c *Context) Serve(val interface{}) {
	c.serveEncoded(val, http.StatusOK, 0)
}

// ServeWithStatus is a helper method to return encoded msg based on type from a struct type.
func (c *Context) ServeWithStatus(val interface{}, status int) {
	c.serveEncoded(val, status, 0)
}

// ServeStatus serves up the status passed in.
func (c *Context) ServeStatus(status int) {
	if status == 0 {
		status = http.StatusOK
	}
	c.Status = status
	c.Response.WriteHeader(c.Status)
}

// ServeCachedWithStatus is a helper method to return encoded msg based on type from a struct type.
func (c *Context) ServeCachedWithStatus(val interface{}, status int, seconds int) {
	c.serveEncoded(val, status, seconds)
}

// serveEncoded serves a value (val) encoded with expiring in seconds and a status
func (c *Context) serveEncoded(val interface{}, status int, seconds int) {
	if status == 0 {
		status = http.StatusOK
	}

	c.Response.Header().Set("Content-Type", c.coder.ContentType())
	if seconds > 0 {
		c.Response.Header().Set(cacheControlHeader, fmt.Sprintf("private, must-revalidate, max-age=%d", seconds))
	}

	c.Response.WriteHeader(status)

	if val != nil {
		if err := c.coder.Encode(c.Response, val); err != nil {
			c.Response.WriteHeader(http.StatusInternalServerError)
			status = http.StatusInternalServerError
		}
	}

	c.Status = status
}

// ServeResponse serves a response with the status and content type sent
func (c *Context) ServeResponse(resp []byte, status int, contentType string) {

	if contentType != "" {
		c.Response.Header().Set("Content-Type", contentType)
	}
	if contentType == "" {
		c.Response.Header().Set("Content-Type", c.coder.ContentType())
	}
	c.Response.WriteHeader(status)
	c.Response.Write(resp)
}
