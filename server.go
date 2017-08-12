package beam

import (
	"net"
	"sync"
	"time"

	"github.com/gaemma/logging"
)

// NewServer creates a redis protocol supported server.
func NewServer(config Config) *Server {
	s := new(Server)
	if config.Timeout <= 0 {
		config.Timeout = time.Second * 30
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

	s.logger.Info("listen on: %s.", addr)

	sleep := time.Second
	for {
		if s.closed() {
			return nil
		}

		conn, err := l.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				time.Sleep(sleep)
				sleep *= 2
				continue
			}
			if s.closed() {
				s.logger.Warning("server is closed already: %s.", err.Error())
				break
			}
			return err
		}
		sleep = time.Second

		s.wg.Add(1)
		go protectCall(func() { s.handleConn(conn) }, s.logger)
	}

	s.wg.Wait()
	return nil
}

// Stop stops the running server.
func (s *Server) Stop() error {
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

func (s *Server) handleConn(conn net.Conn) {
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
	for {
		if s.closed() {
			return
		}
		err := conn.SetReadDeadline(time.Now().Add(s.config.Timeout))
		if err != nil {
			s.logger.Error("fail to set read deadline: %s.", err.Error())
			return
		}

		req, err := ReadRequest(conn)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				s.logger.Warning("timeout from %s.", conn.RemoteAddr())
				return
			}
			s.logger.Error("fail to read request: %s.", err.Error())
			return
		}

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

		err = conn.SetWriteDeadline(time.Now().Add(s.config.Timeout))
		if err != nil {
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
