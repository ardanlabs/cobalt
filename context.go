package cobalt

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	CacheControlHeader = "Cache-control"
)

type (
	// Context is the struct type that holds context data for a request.
	Context struct {
		Response http.ResponseWriter
		Request  *http.Request
		Data     map[interface{}]interface{}
	}
)

// NewContext creates a new context instance with a http.Request and http.ResponseWriter.
func NewContext(req *http.Request, resp http.ResponseWriter) *Context {
	return &Context{Request: req, Response: resp, Data: map[interface{}]interface{}{}}
}

// GetValue returns the value for the associated key from the url parameters.
func (c *Context) GetValue(key string) string {
	return c.Request.URL.Query().Get(key)
}

// GetData returns the value for the specified key from the context data. Usually used by prefilters to pass data to the http handler
// and post filters.
func (c *Context) GetData(key interface{}) interface{} {
	return c.Data[key]
}

// SetData sets the data for the specified key in the context instance.
func (c *Context) SetData(key interface{}, value interface{}) {
	c.Data[key] = value
}

// ServeJson is a helper method to return json from a struct type.
func (c *Context) ServeJson(obj interface{}) {
	c.ServeJsonWithCache(obj, 0)
}

// ServeJsonWith is a helper method to return json from a struct type. It adds a cache control header
// to the response if seconds > 0
func (c *Context) ServeJsonWithCache(obj interface{}, seconds int64) {
	if seconds > 0 {
		c.Response.Header().Set(CacheControlHeader, fmt.Sprintf("private, must-revalidate, max-age=%d", seconds))
	}
	c.Response.Header().Set("Content-Type", "application/json;charset=UTF-8")
	json.NewEncoder(c.Response).Encode(obj)
}

// ServeJsonWithStatus is a helper method to return json from a struct type with a status code.
func (c *Context) ServeJsonWithStatus(status int, obj interface{}) {
	c.Response.Header().Set("Content-Type", "application/json;charset=UTF-8")
	c.Response.WriteHeader(status)
	json.NewEncoder(c.Response).Encode(obj)
}
