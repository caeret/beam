package beam

import (
	"fmt"
	"strings"
)

var (
	crlf = []byte{'\r', '\n'}
)

func escapeCrlf(data string) string {
	return strings.NewReplacer("\r", "\\r", "\n", "\\n").Replace(data)
}

func protectCall(call func(), logger Logger) {
	defer func() {
		if err := recover(); err != nil {
			logger.Log(LogLevelError, fmt.Sprintf("panic: %v.", err))
		}
	}()
	call()
}
