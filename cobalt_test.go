package cobalt

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var r = map[int][]string{
	1: []string{"/", "Get"},
	2: []string{"/foo", "Get"},
	3: []string{"/", "Post"},
	4: []string{"/foo", "Post"},
	5: []string{"/", "Put"},
	6: []string{"/foo", "Put"}}

func newRequest(method, path string, body io.Reader) *http.Request {
	r, _ := http.NewRequest(method, path, body)
	u, _ := url.Parse(path)
	r.URL = u
	r.RequestURI = path
	return r
}

func Test_PreFilters(t *testing.T) {
	//setup request
	r := newRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	data := "PREFILTER"
	// pre filters
	c := New()

	c.AddPrefilter(func(ctx *Context) bool {
		ctx.SetData("PRE", data)
		return false
	})

	c.Get("/", func(ctx *Context) {
		v := ctx.GetData("PRE")
		if v != data {
			t.Errorf("expected %s got %s", data, v)
		}
		ctx.Response.Write([]byte(data))
	}, nil)

	c.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status code to be 200 instead got %d", w.Code)
	}
	if w.Body.String() != data {
		t.Errorf("expected body to be %s instead got %s", w.Body.String())
	}
}

func Test_PreFiltersExit(t *testing.T) {
	r := newRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	data := "PREFILTER_EXIT"
	code := http.StatusBadRequest
	c := New()

	c.AddPrefilter(func(ctx *Context) bool {
		ctx.Response.WriteHeader(code)
		ctx.Response.Write([]byte(data))
		return true
	})

	c.Get("/", func(ctx *Context) {
		v := ctx.GetData("PRE")
		if v != data {
			t.Errorf("expected %s got %s", data, v)
		}
		ctx.Response.Write([]byte(data))
	}, nil)

	c.ServeHTTP(w, r)

	if w.Code != code {
		t.Errorf("expected status code to be 200 instead got %d", w.Code)
	}
	if w.Body.String() != data {
		t.Errorf("expected body to be %s instead got %s", w.Body.String())
	}
}

func Test_Routes(t *testing.T) {
	c := New()

	// GET
	c.Get("/", func(ctx *Context) {
		ctx.Response.Write([]byte("Get/"))
	}, nil)
	c.Get("/foo", func(ctx *Context) {
		ctx.Response.Write([]byte("Get/foo"))
	}, nil)

	// POST
	c.Post("/", func(ctx *Context) {
		ctx.Response.Write([]byte("Post/"))
	}, nil)
	c.Post("/foo", func(ctx *Context) {
		ctx.Response.Write([]byte("Post/foo"))
	}, nil)

	// PUT
	c.Put("/", func(ctx *Context) {
		ctx.Response.Write([]byte("Put/"))
	}, nil)
	c.Put("/foo", func(ctx *Context) {
		ctx.Response.Write([]byte("Put/foo"))
	}, nil)

	// Delete
	c.Delete("/", func(ctx *Context) {
		ctx.Response.Write([]byte("Delete/"))
	}, nil)
	c.Delete("/foo", func(ctx *Context) {
		ctx.Response.Write([]byte("Delete/foo"))
	}, nil)

	// OPTIONS
	c.Options("/", func(ctx *Context) {
		ctx.Response.Write([]byte("Options/"))
	}, nil)
	c.Options("/foo", func(ctx *Context) {
		ctx.Response.Write([]byte("Options/foo"))
	}, nil)

	// HEAD
	c.Head("/", func(ctx *Context) {
		ctx.Response.Write([]byte("Head/"))
	}, nil)
	c.Head("/foo", func(ctx *Context) {
		ctx.Response.Write([]byte("Head/foo"))
	}, nil)

	for _, v := range r {
		AssertRoute(v[0], v[1], c, t)
	}
}

func AssertRoute(path, verb string, c *Cobalt, t *testing.T) {
	r := newRequest(strings.ToUpper(verb), path, nil)
	w := httptest.NewRecorder()

	c.ServeHTTP(w, r)
	if w.Body.String() != verb+path {
		t.Errorf("expected body to be %s instead got %s", w.Body.String())
	}
}

func Test_PostFilters(t *testing.T) {

}
