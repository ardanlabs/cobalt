package helper

import (
	"encoding/json"

	"github.com/goinggo/tracelog"
)

const (
	MAIN_GO_ROUTINE     = "main"
	MONGO_DATABASE      = "outcast"
	EMAIL_ALERT_SUBJECT = "Outcaster Alert"
)

// LogModel logs a model in pretty print json.
func LogModel(obj interface{}, useTrace bool) {
	bArray, _ := json.MarshalIndent(obj, "", "    ")

	if useTrace {
		tracelog.TRACE("utils", "LogModel", "Obj => \n\n%s\n\n", string(bArray))
		return
	}

	tracelog.INFO("utils", "LogModel", "Obj => \n\n%s\n\n", string(bArray))
}
