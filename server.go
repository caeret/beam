package beam

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gaemma/logging"
)

var ErrServerClosed = errors.New("beam: Server closed")

// NewServer creates a redis protocol supported server.
func NewServer(config Config) *Server {
	s := new(Server)
	if config.RWTimeout <= 0 {
		config.RWTimeout = time.Second * 5
	}
	if config.IdleTimeout <= 0 {
		config.IdleTimeout = time.Minute * 5
	}
	s.config = config
	if config.Logger == nil {
		s.logger = logging.NewNopLogger()
	} else {
		s.logger = config.Logger
	}
	s.handler = config.Handler
	s.closeCh = make(chan struct{})
	return s
}

// Server is a redis protocol supported engine.
type Server struct {
	config   Config
	logger   logging.Logger
	handler  Handler
	listener net.Listener
	wg       sync.WaitGroup
	closeCh  chan struct{}
}

// Serve runs the server engine on the given addr.
func (s *Server) Serve(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = l

	s.logger.Info("boot the beam server \"%s\".", addr)

	sleep := time.Second
	for {
		if s.closed() {
			return ErrServerClosed
		}

		var conn net.Conn
		conn, err = l.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				time.Sleep(sleep)
				sleep *= 2
				continue
			}
			if s.closed() {
				err = ErrServerClosed
				break
			}
			return err
		}
		sleep = time.Second

		s.wg.Add(1)
		go protectCall(func() { s.handleConn(NewConn(conn)) }, s.logger)
	}

	s.wg.Wait()
	return err
}

// Stop stops the running server.
func (s *Server) Stop() error {
	s.logger.Info("server is closed.")
	close(s.closeCh)
	err := s.listener.Close()
	return err
}

func (s *Server) closed() bool {
	select {
	case <-s.closeCh:
		return true
	default:
		return false
	}
}

func (s *Server) handleConn(conn *Conn) {
	defer s.wg.Done()

	s.logger.Debug("handle new connection from %s.", conn.RemoteAddr())

	defer func() {
		s.logger.Debug("close connection from %s.", conn.RemoteAddr())
		err := conn.Close()
		if err != nil {
			s.logger.Warning("there is an error when close the connection from %s.", conn.RemoteAddr())
		}
	}()

	var shouldReturn bool
	conn.refreshDeadline(s.config.IdleTimeout)
	for {
		if s.closed() {
			return
		}
		if !conn.beforeDeadline() {
			s.logger.Debug("deadline exceeded from %s.", conn.RemoteAddr())
			return
		}
		err := conn.SetReadDeadline(time.Now().Add(s.config.RWTimeout))
		if err != nil {
			s.logger.Error("fail to set read deadline: %s.", err.Error())
			return
		}

		req, err := ReadRequest(conn)
		if err != nil {
			if err == io.EOF {
				s.logger.Debug("receive EOF from %s.", conn.RemoteAddr())
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				s.logger.Debug("read timeout from %s.", conn.RemoteAddr())
				continue
			}
			s.logger.Error("fail to read request: %s.", err.Error())
			return
		}

		conn.refreshDeadline(s.config.IdleTimeout)

		s.logger.Debug("read request: \"%s\".", req)

		var resp Response

		if s.handler != nil {
			resp, err = s.handler.Handle(req)
			if err != nil {
				shouldReturn = true
				resp = NewErrorsResponse("Error internal error")
				s.logger.Error("fail to handle request: %s", err.Error())
			}
		} else {
			shouldReturn = true
			resp = NewErrorsResponse("Error the Handler should be provided")
		}

		s.logger.Debug("send response: \"%s\".", resp)

		err = conn.SetWriteDeadline(time.Now().Add(s.config.RWTimeout))
		if err != nil {
			if err == io.EOF {
				s.logger.Debug("receive EOF from %s.", conn.RemoteAddr())
				return
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				s.logger.Warning("write timeout from %s.", conn.RemoteAddr())
				return
			}
			s.logger.Error("fail to set write deadline: %s.", err.Error())
			return
		}

		_, err = conn.Write(resp)
		if err != nil {
			shouldReturn = true
			s.logger.Error("fail to write response: %s.", err.Error())
		}

		if shouldReturn {
			return
		}
	}
}
