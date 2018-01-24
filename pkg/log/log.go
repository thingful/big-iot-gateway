package log

import (
	"log"
	"os"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-stack/stack"
)

var logger kitlog.Logger
var df = kitlog.Caller(4)

func init() {
	logger = kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stdout))
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC, "caller", df)
}

// Log wrapper for log
func Log(keyvals ...interface{}) {
	logger.Log(keyvals...)
}

// Fatal
func Fatal(v ...interface{}) {
	log.Fatal(v)
}

func Caller(depth int) func() interface{} {
	return func() interface{} { return stack.Caller(depth) }
}
