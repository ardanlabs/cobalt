// helper package provides generic helper methods to use with the app.
package helper

import (
	"fmt"
	"runtime"

	"github.com/goinggo/tracelog"
)

// CatchPanic calls CatchPanicErr with error as nil is used to catch any Panic and log exceptions to Stdout.
func CatchPanic(sessionId string, functionName string) {
	CatchPanicErr(nil, sessionId, functionName)
}

// CatchPanicErr is used to catch any Panic and log exceptions to Stdout. It will also write the stack trace.
func CatchPanicErr(err *error, sessionId string, functionName string) {
	if r := recover(); r != nil {
		buf := make([]byte, 10000)
		runtime.Stack(buf, false)

		tracelog.ALERT("tracelog.EmailAlertSubject", sessionId, functionName, "PANIC Defered [%v] : Stack Trace : %v", r, string(buf))

		if err != nil {
			*err = fmt.Errorf("%v", r)
		}
	}
}
