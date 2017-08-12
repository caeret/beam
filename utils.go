package beam

import (
	"strings"

	"github.com/gaemma/logging"
)

var (
	crlf = []byte{'\r', '\n'}
)

func escapeCrlf(data string) string {
	return strings.NewReplacer("\r", "\\r", "\n", "\\n").Replace(data)
}

func protectCall(call func(), logger logging.Logger) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("recover from panic: %v.", err)
		}
	}()
	call()
}
