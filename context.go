package cobalt

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// CacheControlHeader represents the http cache control header
	CacheControlHeader = "Cache-control"
)

type (
	// Context is the struct type that holds context data for a request.
	Context struct {
		Response http.ResponseWriter
		Request  *http.Request
		data     map[string]interface{}
		params   map[string]string
	}
)

// NewContext creates a new context instance with a http.Request and http.ResponseWriter.
func NewContext(req *http.Request, resp http.ResponseWriter, p map[string]string) *Context {
	return &Context{
		Request:  req,
		Response: resp,
		data:     make(map[string]interface{}),
		params:   p,
	}
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
	c.Response.Header().Set("Content-Type", "application/json;charset=UTF-8")
	json.NewEncoder(c.Response).Encode(val)
}

// ServeJSONWithStatus is a helper method to return json from a struct type with a status code.
func (c *Context) ServeJSONWithStatus(status int, val interface{}) {
	c.Response.Header().Set("Content-Type", "application/json;charset=UTF-8")
	c.Response.WriteHeader(status)
	json.NewEncoder(c.Response).Encode(val)
}

// ServeJSONString is a helper method to return json from a struct type with a status code.
func (c *Context) ServeJSONString(j string) {
	c.Response.Header().Set("Content-Type", "application/json;charset=UTF-8")
	c.Response.Write([]byte(j))
}
