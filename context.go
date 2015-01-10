package cobalt

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/ardanlabs/msgpack"
)

const (
	// CacheControlHeader represents the http cache control header
	CacheControlHeader = "Cache-control"

	jsonContent    = "application/json;charset=UTF-8"
	msgPackContent = "application/x-msgpack"
)

type (

	// Context is the struct type that holds context data for a request.
	Context struct {
		Response  http.ResponseWriter
		Request   *http.Request
		data      map[string]interface{}
		params    map[string]string
		encodingT EncodingType
	}
)

// NewContext creates a new context instance with a http.Request and http.ResponseWriter.
func NewContext(req *http.Request, resp http.ResponseWriter, p map[string]string, et EncodingType) *Context {
	return &Context{
		Request:   req,
		Response:  resp,
		data:      make(map[string]interface{}),
		params:    p,
		encodingT: et,
	}
}

// Encoding returns the encoding for the context
func (c *Context) Encoding() EncodingType {
	return c.encodingT
}

// RouteValue returns the value for the associated key from the url parameters.
func (c *Context) RouteValue(key string) string {
	return c.params[key]
}

// AllRouteValues returns all the route values.
func (c *Context) AllRouteValues() map[string]string {
	return c.params
}

// GetData returns the value for the specified key from the context data. Usually used by prefilters to pass data to the http handler
// and post filters.
func (c *Context) GetData(key string) interface{} {
	return c.data[key]
}

// SetData sets the data for the specified key in the context instance.
func (c *Context) SetData(key string, value interface{}) {
	c.data[key] = value
}

// Error returns an http Error with the specified Error string and code
func (c *Context) Error(body interface{}, status int) {
	c.ServeWithStatus(status, body)
}

// Serve is a helper method to return encoded msg based on type from a struct type.
func (c *Context) Serve(val interface{}) {
	if c.encodingT == MSGPackEncoding {
		c.ServeMPackWithCache(val, 0)
		return
	}

	c.ServeJSONWithCache(val, 0)
}

// ServeWithCache serves msg with cache length encoded with current encoding
func (c *Context) ServeWithCache(val interface{}, seconds int64) {
	if c.encodingT == MSGPackEncoding {
		c.ServeMPackWithCache(val, 0)
		return
	}

	c.ServeJSONWithCache(val, 0)
}

// ServeWithStatus is a helper method to return encoded response from a struct type with a status code.
func (c *Context) ServeWithStatus(status int, val interface{}) {
	if c.encodingT == MSGPackEncoding {
		c.ServeMPackWithStatus(status, val)
		return
	}

	c.ServeJSONWithStatus(status, val)
}

// ServeJSON is a helper method to return json from a struct type.
func (c *Context) ServeJSON(val interface{}) {
	c.ServeJSONWithCache(val, 0)
}

// ServeJSONWithCache is a helper method to return json from a struct type. It adds a cache control header
// to the response if seconds > 0
func (c *Context) ServeJSONWithCache(val interface{}, seconds int64) {
	if seconds > 0 {
		c.Response.Header().Set(CacheControlHeader, fmt.Sprintf("private, must-revalidate, max-age=%d", seconds))
	}
	c.Response.Header().Set("Content-Type", jsonContent)
	json.NewEncoder(c.Response).Encode(val)
}

// ServeJSONWithStatus is a helper method to return json from a struct type with a status code.
func (c *Context) ServeJSONWithStatus(status int, val interface{}) {
	c.Response.Header().Set("Content-Type", jsonContent)
	c.Response.WriteHeader(status)
	json.NewEncoder(c.Response).Encode(val)
}

// ServeJSONString is a helper method to return json from a struct type with a status code.
func (c *Context) ServeJSONString(j string) {
	c.Response.Header().Set("Content-Type", jsonContent)
	c.Response.Write([]byte(j))
}

// ServeResponse serves a response with the status and content type sent
func (c *Context) ServeResponse(resp []byte, status int, contentType string) {
	if contentType != "" {
		c.Response.Header().Set("Content-Type", contentType)
	}
	c.Response.WriteHeader(status)
	c.Response.Write(resp)
}

// ServeMPack is a helper method to return msgpack binary from a struct type.
func (c *Context) ServeMPack(val interface{}) {
	c.ServeMPackWithCache(val, 0)
}

// ServeMPackWithCache is a helper method to return msgpack binary from a struct type. It adds a cache control header
// to the response if seconds > 0
func (c *Context) ServeMPackWithCache(val interface{}, seconds int64) {
	if seconds > 0 {
		c.Response.Header().Set(CacheControlHeader, fmt.Sprintf("private, must-revalidate, max-age=%d", seconds))
	}
	c.Response.Header().Set("Content-Type", msgPackContent)
	msgpack.NewEncoder(c.Response).Encode(val)
}

// ServeMPackWithStatus is a helper method to return msgpack from a struct type with a status code.
func (c *Context) ServeMPackWithStatus(status int, val interface{}) {
	c.Response.Header().Set("Content-Type", msgPackContent)
	c.Response.WriteHeader(status)
	msgpack.NewEncoder(c.Response).Encode(val)
}
