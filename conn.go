package beam

import (
	"net"
	"time"
)

// NewConn creates a wrapped connection.
func NewConn(conn net.Conn) *Conn {
	c := new(Conn)
	c.Conn = conn
	return c
}

// Conn contains the client connection and deadline for closing.
type Conn struct {
	net.Conn
	deadline time.Time
}

// refreshDeadline sets the deadline with the current time and the given duration d.
func (c *Conn) refreshDeadline(d time.Duration) {
	c.deadline = time.Now().Add(d)
}

// beforeDeadline checks if the connection before the deadline.
func (c *Conn) beforeDeadline() bool {
	return time.Now().Before(c.deadline)
}
