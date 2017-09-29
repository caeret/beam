package beam

import (
	"time"

	"github.com/gaemma/logging"
)

// Config provides the configuration needs by the server.
type Config struct {
	Logger      logging.Logger
	RWTimeout   time.Duration
	IdleTimeout time.Duration
	BufferSize  int
	Network     string
	Addr        string
}
