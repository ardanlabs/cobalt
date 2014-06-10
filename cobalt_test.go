package cobalt

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var r map[int][]string = map[int][]string{
	1: []string{"/", "Get"},
	2: []string{"/foo", "Get"},
	3: []string{"/", "Post"},
	4: []string{"/foo", "Post"},
	5: []string{"/", "Put"},
	6: []string{"/foo", "Put"}}

func Test_PreFilters(t *testing.T) {
	//setup request
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	data := "PREFILTER"
	// pre filters
	d := New()

	d.AddPrefilter(func(c *Context) bool {
		c.SetData("PRE", data)
		return false
	})

	d.Get("/", func(ctx *Context) {
		v := ctx.GetData("PRE")
		if v != data {
			t.Errorf("expected %s got %s", data, v)
		}
		ctx.Response.Write([]byte(data))
	})

	d.router.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status code to be 200 instead got %d", w.Code)
	}
	if w.Body.String() != data {
		t.Errorf("expected body to be %s instead got %s", w.Body.String())
	}
}

func Test_PreFiltersExit(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	data := "PREFILTER_EXIT"
	code := http.StatusBadRequest
	d := New()

	d.AddPrefilter(func(c *Context) bool {
		c.Response.WriteHeader(code)
		c.Response.Write([]byte(data))
		return true
	})

	d.Get("/", func(ctx *Context) {
		v := ctx.GetData("PRE")
		if v != data {
			t.Errorf("expected %s got %s", data, v)
		}
		ctx.Response.Write([]byte(data))
	})

	d.router.ServeHTTP(w, r)

	if w.Code != code {
		t.Errorf("expected status code to be 200 instead got %d", w.Code)
	}
	if w.Body.String() != data {
		t.Errorf("expected body to be %s instead got %s", w.Body.String())
	}
}

func Test_Routes(t *testing.T) {
	d := New()

	// GET
	d.Get("/", func(c *Context) {
		c.Response.Write([]byte("Get/"))
	})
	d.Get("/foo", func(c *Context) {
		c.Response.Write([]byte("Get/foo"))
	})

	// POST
	d.Post("/", func(c *Context) {
		c.Response.Write([]byte("Post/"))
	})
	d.Post("/foo", func(c *Context) {
		c.Response.Write([]byte("Post/foo"))
	})

	// PUT
	d.Put("/", func(c *Context) {
		c.Response.Write([]byte("Put/"))
	})
	d.Put("/foo", func(c *Context) {
		c.Response.Write([]byte("Put/foo"))
	})

	// Delete
	d.Delete("/", func(c *Context) {
		c.Response.Write([]byte("Delete/"))
	})
	d.Delete("/foo", func(c *Context) {
		c.Response.Write([]byte("Delete/foo"))
	})

	// OPTIONS
	d.Options("/", func(c *Context) {
		c.Response.Write([]byte("Options/"))
	})
	d.Options("/foo", func(c *Context) {
		c.Response.Write([]byte("Options/foo"))
	})

	// HEAD
	d.Head("/", func(c *Context) {
		c.Response.Write([]byte("Head/"))
	})
	d.Head("/foo", func(c *Context) {
		c.Response.Write([]byte("Head/foo"))
	})

	for _, v := range r {
		AssertRoute(v[0], v[1], d, t)
	}
}

func AssertRoute(path, verb string, d *Dispatcher, t *testing.T) {
	r, _ := http.NewRequest(strings.ToUpper(verb), path, nil)
	w := httptest.NewRecorder()

	d.router.ServeHTTP(w, r)
	if w.Body.String() != verb+path {
		t.Errorf("expected body to be %s instead got %s", w.Body.String())
	}
}

func Test_PostFilters(t *testing.T) {

}
