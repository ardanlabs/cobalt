package cobalt

import (
	"encoding/json"
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

// Test_PreFilters tests pre-filters
func Test_PreFilters(t *testing.T) {
	//setup request
	r := newRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	data := "PREFILTER"
	// pre filters
	c := New(JSONEncoding)

	c.AddPrefilter(func(ctx *Context) bool {
		ctx.SetData("PRE", data)
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

	if w.Code != 200 {
		t.Errorf("expected status code to be 200 instead got %d", w.Code)
	}
	if w.Body.String() != data {
		t.Errorf("expected body to be %s instead got %s", data, w.Body.String())
	}
}

// Test_PreFiltersExit tests pre-filters stopping the request.
func Test_PreFiltersExit(t *testing.T) {
	r := newRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	data := "PREFILTER_EXIT"
	code := http.StatusBadRequest
	c := New(JSONEncoding)

	c.AddPrefilter(func(ctx *Context) bool {
		ctx.Response.WriteHeader(code)
		ctx.Response.Write([]byte(data))
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

	if w.Code != code {
		t.Errorf("expected status code to be 200 instead got %d", w.Code)
	}
	if w.Body.String() != data {
		t.Errorf("expected body to be %s instead got %s", data, w.Body.String())
	}
}

// Test_Routes tests the routing of requests.
func Test_Routes(t *testing.T) {
	c := New(JSONEncoding)

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

// Test_RouteFiltersSettingData tests route filters setting data and passing it to handlers.
func Test_RouteFiltersSettingData(t *testing.T) {

	//setup request
	r := newRequest("GET", "/RouteFilter", nil)
	w := httptest.NewRecorder()

	// test route filter setting
	data := "ROUTEFILTER"

	c := New(JSONEncoding)

	c.Get("/RouteFilter",

		func(ctx *Context) {
			v := ctx.GetData("PRE")
			if v != data {
				t.Errorf("expected %s got %s", data, v)
			}
			ctx.Response.Write([]byte(data))
		},
		[]FilterHandler{
			func(c *Context) bool {
				c.SetData("PRE", data)
				return true
			}})

	c.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected status code to be 200 instead got %d", w.Code)
	}
	if w.Body.String() != data {
		t.Errorf("expected body to be %s instead got %s", data, w.Body.String())
	}
}

// Test_RouteFilterExit tests route filters stopping the request.
func Test_RouteFilterExit(t *testing.T) {
	data := "ROUTEFILTEREXIT"
	//setup request
	r := newRequest("GET", "/RouteFilter", nil)
	w := httptest.NewRecorder()

	c := New(JSONEncoding)

	c.Get("/RouteFilter",

		func(ctx *Context) {
			v := ctx.GetData("PRE")
			if v != data {
				t.Errorf("expected %s got %s", data, v)
			}
			ctx.Response.Write([]byte("FOO"))
		},
		[]FilterHandler{
			func(ctx *Context) bool {
				ctx.Response.WriteHeader(http.StatusUnauthorized)
				ctx.Response.Write([]byte(data))
				return false
			}})

	c.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status code to be %d instead got %d", http.StatusUnauthorized, w.Code)
	}
	if w.Body.String() != data {
		t.Errorf("expected body to be %s instead got %s", data, w.Body.String())
	}
}

// AsserRoute is a helper method to tests routes
func AssertRoute(path, verb string, c *Cobalt, t *testing.T) {
	r := newRequest(strings.ToUpper(verb), path, nil)
	w := httptest.NewRecorder()

	c.ServeHTTP(w, r)
	if w.Body.String() != verb+path {
		t.Errorf("expected body to be %s instead got %s", verb+path, w.Body.String())
	}
}

func Test_PostFilters(t *testing.T) {

}

func Test_NotFoundHandler(t *testing.T) {
	//setup request
	r := newRequest("GET", "/FOO", nil)
	w := httptest.NewRecorder()

	m := struct{ Message string }{"Not Found"}

	nf := func(c *Context) {
		c.ServeJSONWithStatus(http.StatusNotFound, m)
	}

	c := New(JSONEncoding)
	c.AddNotFoundHandler(nf)

	c.Get("/",
		func(ctx *Context) {
			panic("Panic Test")
		},
		nil)

	c.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status code to be 404 instead got %d", w.Code)
	}

	var msg struct{ Message string }
	json.Unmarshal([]byte(w.Body.String()), &msg)

	if msg.Message != m.Message {
		t.Errorf("expected body to be %s instead got %s", msg.Message, m.Message)
	}
}

func Test_ServerErrorHandler(t *testing.T) {
	//setup request
	r := newRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	m := struct{ Message string }{"Internal Error"}

	se := func(c *Context) {
		c.ServeJSONWithStatus(http.StatusInternalServerError, m)
	}

	c := New(JSONEncoding)
	c.AddServerErrHanlder(se)

	c.Get("/",
		func(ctx *Context) {
			panic("Panic Test")
		},
		nil)

	c.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status code to be 500 instead got %d", w.Code)
	}

	var msg struct{ Message string }
	json.Unmarshal([]byte(w.Body.String()), &msg)

	if msg.Message != m.Message {
		t.Errorf("expected body to be %s instead got %s", msg.Message, m.Message)
	}
}
