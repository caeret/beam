package beam

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/gaemma/logging"
)

// ErrServerClosed will be returned when beam server is closed.
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
		go protectCall(s.createClient(conn, 16*1024, func() {
			s.wg.Done()
		}).run, s.logger)
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

func (s *Server) createClient(conn net.Conn, bufferSize int, closeFunc func()) *Client {
	c := new(Client)
	c.logger = s.logger
	c.conn = conn
	c.b = make([]byte, bufferSize)
	c.bsize = bufferSize
	c.rwTimeout = s.config.RWTimeout
	c.idleTimeout = s.config.IdleTimeout
	c.handler = s.config.Handler
	c.closeFun = closeFunc
	c.closeCh = s.closeCh
	c.stats = new(ClientStats)
	return c
}
