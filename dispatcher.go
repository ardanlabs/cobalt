package cobalt

import (
	"fmt"
	"net/http"

	"github.com/OutCast-IO/cobalt/utilities/mongo"
	"github.com/codegangsta/negroni"
	"github.com/goinggo/tracelog"
	"github.com/gorilla/mux"
)

const RequestCtx key = 0

type (
	key int

	Dispatcher struct {
		Router *mux.Router
		N      *negroni.Negroni
	}

	Handler    func(c *Ctx)
	Middleware interface {
		negroni.Handler
	}
)

func NewDispatcher() *Dispatcher {
	r := mux.NewRouter()
	n := negroni.New()
	return &Dispatcher{r, n}
}

func (d *Dispatcher) AddRoute(route string, h Handler) {
	d.Router.HandleFunc(route, func(w http.ResponseWriter, req *http.Request) {
		var c Ctx
		var err error
		c.Session, err = mongo.CopyMonotonicSession("Dispatcher Add Route Handler")
		if err != nil {
			tracelog.ERRORf(err, "", "Dispatcher Add Route Handler, Route %s", route)
			fmt.Fprintf(w, "Abort 500, No Mongo")
			return
		}
		defer c.Session.Close()

		c.UUID = "User Id"
		c.Request = req
		c.Response = w

		h(&c)
	})
}

func (d *Dispatcher) Run(addr string) {
	d.N.UseHandler(d.Router)
	d.N.Run(addr)
}

func (d *Dispatcher) Use(m Middleware) {
	d.N.Use(m)
}
