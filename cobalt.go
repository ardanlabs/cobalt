package cobalt

import (
	"net/http"

	"github.com/goinggo/tracelog"
	"labix.org/v2/mgo"
)

type (
	App struct{}

	Ctx struct {
		Session  *mgo.Session
		UUID     string
		Response http.ResponseWriter
		Request  *http.Request
	}

	Logger struct{}
)

func (l Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	tracelog.COMPLETED("pluma", "ServeHTTP")

	next(rw, r)

	tracelog.COMPLETED("pluma", "ServeHTTP completed")
}

func (a App) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	tracelog.STARTED("pluma", "App.ServeHTTP")

	next(rw, r)

	tracelog.COMPLETED("pluma", "App.ServeHTTP")
}
