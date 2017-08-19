package beam

import (
	"io"
	"net"
	"time"

	"github.com/gaemma/logging"
)

// NewRequest creates a new Request with the given Client and Request.
func NewRequest(client *Client, cmd Command) *Request {
	req := new(Request)
	req.Client = client
	req.Command = cmd
	return req
}

// Request links the Request with the Client.
type Request struct {
	*Client
	Command
}

// ClientStats contains the statistics data.
type ClientStats struct {
	BytesIn  int
	BytesOut int
	Commands int
}

// Client contains the client connection and deadline for closing.
type Client struct {
	logger      logging.Logger
	conn        net.Conn
	deadline    time.Time
	b           []byte
	bsize       int
	boff        int
	rwTimeout   time.Duration
	idleTimeout time.Duration
	handler     Handler
	closeFun    func(c *Client)
	closeCh     <-chan struct{}
	stats       *ClientStats
}

// Stats retrieves the ClientStats value.
func (c *Client) Stats() ClientStats {
	return *c.stats
}

// refreshDeadline sets the deadline with the current time and the given duration d.
func (c *Client) refreshDeadline(d time.Duration) {
	c.deadline = time.Now().Add(d)
}

// beforeDeadline checks if the connection before the deadline.
func (c *Client) beforeDeadline() bool {
	return time.Now().Before(c.deadline)
}

func (c *Client) run() {
	if c.closeFun != nil {
		defer c.closeFun(c)
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

		c.stats.BytesIn += nr

		var cmds Commands
		l := c.b[:c.boff+nr]
		for {
			var cmd Command
			cmd, l, err = ReadCommand(l)
			if err != nil {
				if err == io.EOF {
					copy(c.b, l)
					c.boff = len(l)
					break
				}
				c.logger.Error("fail to read request: %s.", err.Error())
				return
			}
			cmds = append(cmds, cmd)
		}

		c.stats.Commands += len(cmds)
		c.refreshDeadline(c.idleTimeout)

		c.logger.Debug("read %d requests: \"%s\".", len(cmds), cmds)

		var resps Responses

		for _, cmd := range cmds {
			var resp Response

			if c.handler != nil {
				resp, err = c.handler.Handle(NewRequest(c, cmd))
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
		c.stats.BytesOut += nw

		if shouldReturn {
			return
		}
	}
}
