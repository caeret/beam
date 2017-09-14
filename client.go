package beam

import (
	"errors"
	"io"
	"net"
	"time"
)

var (
	ErrHaltClient = errors.New("halt client")
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
	s          *Server
	conn       net.Conn
	deadline   time.Time
	b          []byte
	bsize      int
	boff       int
	stats      *ClientStats
	closeCh    chan struct{}
	attributes map[string]interface{}
}

func (c *Client) GetAttr(key string) interface{} {
	return c.attributes[key]
}

func (c *Client) SetAttr(key string, value interface{}) {
	c.attributes[key] = value
}

func (c *Client) HasAttr(key string) bool {
	_, exist := c.attributes[key]
	return exist
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
	defer c.s.stopClient(c)

	c.s.logger.Debug("handle new connection from %s.", c.conn.RemoteAddr())

	defer func() {
		c.s.logger.Debug("close connection from %s.", c.conn.RemoteAddr())
		err := c.conn.Close()
		if err != nil {
			c.s.logger.Warning("there is an error when close the connection from %s.", c.conn.RemoteAddr())
		}
	}()

	var shouldReturn bool
	c.refreshDeadline(c.s.config.IdleTimeout)

	for {
		select {
		case <-c.s.closeCh:
			return
		case <-c.closeCh:
			return
		default:
		}
		if !c.beforeDeadline() {
			c.s.logger.Debug("deadline exceeded from %s.", c.conn.RemoteAddr())
			return
		}
		err := c.conn.SetReadDeadline(time.Now().Add(c.s.config.RWTimeout))
		if err != nil {
			c.s.logger.Error("fail to set read deadline: %s.", err.Error())
			return
		}

		nr, err := c.conn.Read(c.b[c.boff:])
		if err != nil {
			if err == io.EOF {
				c.s.logger.Debug("receive EOF from %s.", c.conn.RemoteAddr())
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				c.s.logger.Debug("read timeout from %s.", c.conn.RemoteAddr())
				continue
			}
			c.s.logger.Error("fail to read request: %s.", err.Error())
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
				c.s.logger.Error("fail to read request: %s.", err.Error())
				return
			}
			cmds = append(cmds, cmd)
		}

		c.stats.Commands += len(cmds)
		c.refreshDeadline(c.s.config.IdleTimeout)

		c.s.logger.Debug("read %d requests: \"%s\".", len(cmds), cmds)

		var replies Replies

		for _, cmd := range cmds {
			var reply Reply

			if c.s.handler != nil {
				reply, err = c.s.handler.Handle(NewRequest(c, cmd))
				if err != nil {
					if err == ErrHaltClient {
						shouldReturn = true
						reply = NewErrorsReply("ERR connection is closed by the server")
					} else {
						reply = NewErrorsReply("ERR internal server error")
						c.s.logger.Error("fail to handle request: %s", err.Error())
					}
				}
			} else {
				reply = NewErrorsReply("ERR a request handler should be provided")
			}

			replies = append(replies, reply)
		}

		c.s.logger.Debug("send %d responses: \"%s\".", len(replies), replies)

		err = c.conn.SetWriteDeadline(time.Now().Add(c.s.config.RWTimeout))
		if err != nil {
			if err == io.EOF {
				c.s.logger.Debug("receive EOF from %s.", c.conn.RemoteAddr())
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				c.s.logger.Warning("write timeout from %s.", c.conn.RemoteAddr())
				return
			}
			c.s.logger.Error("fail to set write deadline: %s.", err.Error())
			return
		}

		nw, err := c.conn.Write([]byte(replies.Raw()))
		if err != nil {
			shouldReturn = true
			c.s.logger.Error("fail to write response: %s.", err.Error())
		}
		c.stats.BytesOut += nw

		if shouldReturn {
			return
		}
	}
}

func (c *Client) stop() {
	select {
	case <-c.closeCh:
		c.s.logger.Debug("client is already closed: %s.", c.conn.RemoteAddr)
	default:
		c.closeCh <- struct{}{}
	}
}
