package beam

import (
	"io"
	"net"
	"time"

	"github.com/gaemma/logging"
)

// Conn contains the client connection and deadline for closing.
type client struct {
	logger      logging.Logger
	conn        net.Conn
	deadline    time.Time
	b           []byte
	bsize       int
	boff        int
	bytesIn     int
	bytesOut    int
	rwTimeout   time.Duration
	idleTimeout time.Duration
	handler     Handler
	closeFun    func()
	closeCh     <-chan struct{}
}

// refreshDeadline sets the deadline with the current time and the given duration d.
func (c *client) refreshDeadline(d time.Duration) {
	c.deadline = time.Now().Add(d)
}

// beforeDeadline checks if the connection before the deadline.
func (c *client) beforeDeadline() bool {
	return time.Now().Before(c.deadline)
}

func (c *client) run() {
	if c.closeFun != nil {
		defer c.closeFun()
	}

	c.logger.Debug("handle new connection from %s.", c.conn.RemoteAddr())

	defer func() {
		c.logger.Debug("close connection from %s.", c.conn.RemoteAddr())
		err := c.conn.Close()
		if err != nil {
			c.logger.Warning("there is an error when close the connection from %s.", c.conn.RemoteAddr())
		}
	}()

	var shouldReturn bool
	c.refreshDeadline(c.idleTimeout)

	for {
		select {
		case <-c.closeCh:
			return
		default:
		}
		if !c.beforeDeadline() {
			c.logger.Debug("deadline exceeded from %s.", c.conn.RemoteAddr())
			return
		}
		err := c.conn.SetReadDeadline(time.Now().Add(c.rwTimeout))
		if err != nil {
			c.logger.Error("fail to set read deadline: %s.", err.Error())
			return
		}

		nr, err := c.conn.Read(c.b[c.boff:])
		if err != nil {
			if err == io.EOF {
				c.logger.Debug("receive EOF from %s.", c.conn.RemoteAddr())
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				c.logger.Debug("read timeout from %s.", c.conn.RemoteAddr())
				continue
			}
			c.logger.Error("fail to read request: %s.", err.Error())
			return
		}

		c.bytesIn += nr

		var reqs Requests
		l := c.b[:c.boff+nr]
		for {
			var req Request
			req, l, err = ReadRequest(l)
			if err != nil {
				if err == io.EOF {
					copy(c.b, l)
					c.boff = len(l)
					break
				}
				c.logger.Error("fail to read request: %s.", err.Error())
				return
			}
			reqs = append(reqs, req)
		}

		c.refreshDeadline(c.idleTimeout)

		c.logger.Debug("read %d requests: \"%s\".", len(reqs), reqs)

		var resps Responses

		for _, req := range reqs {
			var resp Response

			if c.handler != nil {
				resp, err = c.handler.Handle(req)
				if err != nil {
					shouldReturn = true
					resp = NewErrorsResponse("Error internal error")
					c.logger.Error("fail to handle request: %s", err.Error())
				}
			} else {
				shouldReturn = true
				resp = NewErrorsResponse("Error the Handler should be provided")
			}

			resps = append(resps, resp)
		}

		c.logger.Debug("send %d responses: \"%s\".", len(resps), resps)

		err = c.conn.SetWriteDeadline(time.Now().Add(c.rwTimeout))
		if err != nil {
			if err == io.EOF {
				c.logger.Debug("receive EOF from %s.", c.conn.RemoteAddr())
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				c.logger.Warning("write timeout from %s.", c.conn.RemoteAddr())
				return
			}
			c.logger.Error("fail to set write deadline: %s.", err.Error())
			return
		}

		nw, err := c.conn.Write([]byte(resps.Raw()))
		if err != nil {
			shouldReturn = true
			c.logger.Error("fail to write response: %s.", err.Error())
		}
		c.bytesOut += nw

		if shouldReturn {
			return
		}
	}
}
