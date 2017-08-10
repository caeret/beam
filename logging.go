package beam

import (
	"log"
)

type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo             = "info"
	LogLevelWarning          = "warning"
	LogLevelError            = "error"
)

var nopLogger = LogFunc(func(level LogLevel, msg string) {})

// SimpleLogger implements the Logger interface which records the logs by the std log lib.
var SimpleLogger = LogFunc(func(level LogLevel, msg string) {
	log.Printf("[%s] %s", level, msg)
})

type LogFunc func(level LogLevel, msg string)

func (lf LogFunc) Log(level LogLevel, msg string) {
	lf(level, msg)
}

// Logger logs the engine messages.
type Logger interface {
	Log(level LogLevel, msg string)
}
