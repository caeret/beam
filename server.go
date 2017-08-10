package beam

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// NewServer creates a redis protocol supported server.
func NewServer(config Config) *Server {
	s := new(Server)
	if config.Timeout <= 0 {
		config.Timeout = time.Second * 30
	}
	s.config = config
	if config.Logger == nil {
		s.logger = nopLogger
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
	logger   Logger
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

	s.logger.Log(LogLevelInfo, fmt.Sprintf("listen on: %s.", addr))

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
				s.logger.Log(LogLevelWarning, fmt.Sprintf("server is closed already: %s.", err.Error()))
				return nil
			}
			return err
		}
		sleep = time.Second

		s.wg.Add(1)
		go protectCall(func() { s.handleConn(conn) }, s.logger)
	}
}

// Stop stops the running server.
func (s *Server) Stop() error {
	close(s.closeCh)
	s.wg.Wait()
	return s.listener.Close()
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

	s.logger.Log(LogLevelDebug, fmt.Sprintf("handle new connection from %s.", conn.RemoteAddr()))

	defer func() {
		s.logger.Log(LogLevelDebug, fmt.Sprintf("close connection from %s.", conn.RemoteAddr()))
		err := conn.Close()
		if err != nil {
			s.logger.Log(LogLevelWarning, fmt.Sprintf("there is an error when close the connection from %s.", conn.RemoteAddr()))
		}
	}()

	var shouldReturn bool
	for {
		if s.closed() {
			return
		}
		err := conn.SetReadDeadline(time.Now().Add(s.config.Timeout))
		if err != nil {
			s.logger.Log(LogLevelError, fmt.Sprintf("fail to set read deadline: %s.", err.Error()))
			return
		}

		req, err := ReadRequest(conn)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				s.logger.Log(LogLevelWarning, fmt.Sprintf("timeout from %s.", conn.RemoteAddr()))
				return
			}
			s.logger.Log(LogLevelError, fmt.Sprintf("fail to read request: %s.", err.Error()))
			return
		}

		s.logger.Log(LogLevelDebug, fmt.Sprintf("read request: \"%s\".", req))

		var resp Response

		if s.handler != nil {
			resp, err = s.handler.Handle(req)
			if err != nil {
				shouldReturn = true
				resp = NewErrorsResponse("Error internal error")
				s.logger.Log(LogLevelError, err.Error())
			}
		} else {
			shouldReturn = true
			resp = NewErrorsResponse("Error the Handler should be provided")
		}

		s.logger.Log(LogLevelDebug, fmt.Sprintf("send response: \"%s\".", resp))

		err = conn.SetWriteDeadline(time.Now().Add(s.config.Timeout))
		if err != nil {
			s.logger.Log(LogLevelError, fmt.Sprintf("fail to set write deadline: %s.", err.Error()))
			return
		}

		_, err = conn.Write(resp)
		if err != nil {
			shouldReturn = true
			s.logger.Log(LogLevelError, fmt.Sprintf("fail to write response: %s.", err.Error()))
		}

		if shouldReturn {
			return
		}
	}
}
