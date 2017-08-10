package beam

import "time"

// Config provides the configuration needs by the server.
type Config struct {
	Logger  Logger
	Handler Handler
	Timeout time.Duration
}
