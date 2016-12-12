package cobalt_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ardanlabs/cobalt"
)

type (
	T1 struct {
		Name   string
		Ti     time.Time
		Amount float64
		Qty    int
		Is     bool
	}
)

func Test_ContextServeJSON(t *testing.T) {
	//setup request
	r := NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	const d = "Jan 5, 2015"
	const dateForm = "Jan 2, 2006"
	ttime, _ := time.Parse(dateForm, d)
	name := "Test JSON"
	amt := 34.56
	qty := 12
	is := true

	t1 := T1{
		Name:   name,
		Ti:     ttime,
		Amount: amt,
		Qty:    qty,
		Is:     is,
	}

	c := cobalt.New(JSONEncoder{})

	c.Get("/", func(c *cobalt.Context) {
		c.Serve(&t1)
	})

	c.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code to be %d instead got %d", http.StatusOK, w.Code)
	}

	if contentType := w.Header().Get("Content-Type"); contentType == "" {
		t.Fatalf("expected content type to not be empty")
	}

	var response T1
	if e := json.Unmarshal([]byte(w.Body.String()), &response); e != nil {
		t.Fatalf("expected no err unmarshaling response, instead got [%s]", e.Error())
	}

	if response.Name != name {
		t.Fatalf("expected name to be %s instead got %s", name, response.Name)
	}
	if response.Ti.Unix() != ttime.Unix() {
		t.Fatalf("expected name to be %s instead got %s", ttime, response.Ti)
	}
	if response.Amount != amt {
		t.Fatalf("expected name to be %f instead got %f", amt, response.Amount)
	}
	if response.Qty != qty {
		t.Fatalf("expected name to be %d instead got %d", qty, response.Qty)
	}

	if response.Is != is {
		t.Fatalf("expected name to be %t instead got %t", is, response.Is)
	}
}
