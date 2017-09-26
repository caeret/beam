package beam

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/gaemma/logging"
)

const defaultBufferSize = 16 * 1024

// ErrServerClosed will be returned when beam server is closed.
var ErrServerClosed = errors.New("beam: Server closed")

// NewServer creates a redis protocol supported server.
func NewServer(handler Handler, config Config) *Server {
	if handler == nil {
		panic(errors.New("a handler should be provided"))
	}
	if config.RWTimeout <= 0 {
		config.RWTimeout = time.Second * 5
	}
	if config.IdleTimeout <= 0 {
		config.IdleTimeout = time.Minute * 5
	}
	if config.BufferSize <= 0 {
		config.BufferSize = defaultBufferSize
	}
	s := new(Server)
	s.handler = handler
	s.config = config
	if config.Logger == nil {
		s.logger = logging.NewNopLogger()
	} else {
		s.logger = config.Logger
	}
	s.closeCh = make(chan struct{})
	s.clients = make(map[string]*Client)
	return s
}

// Server is a redis protocol supported engine.
type Server struct {
	config       Config
	logger       logging.Logger
	handler      Handler
	listener     net.Listener
	clientsWait  sync.WaitGroup
	closeCh      chan struct{}
	clients      map[string]*Client
	clientsMutex sync.RWMutex
}

// Serve runs the server engine on the given addr. if beam server is closed, ErrServerClosed will be retuend.
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

		client := s.createClient(conn, s.config.BufferSize)
		s.startClient(client)
	}

	s.clientsWait.Wait()
	return err
}

func (s *Server) startClient(client *Client) {
	s.clientsWait.Add(1)
	s.clientsMutex.Lock()
	s.clients[client.conn.RemoteAddr().String()] = client
	s.clientsMutex.Unlock()
	go protectCall(client.run, s.logger)
}

func (s *Server) stopClient(client *Client) {
	s.clientsWait.Done()
	s.clientsMutex.Lock()
	delete(s.clients, client.conn.RemoteAddr().String())
	s.clientsMutex.Unlock()
}

// Close stops the running server.
func (s *Server) Close() error {
	select {
	case <-s.closeCh:
		return nil
	default:
		s.logger.Info("server is closed.")
		close(s.closeCh)
		err := s.listener.Close()
		return err
	}
}

func (s *Server) closed() bool {
	select {
	case <-s.closeCh:
		return true
	default:
		return false
	}
}

func (s *Server) createClient(conn net.Conn, bufferSize int) *Client {
	c := new(Client)
	c.s = s
	c.conn = conn
	c.b = make([]byte, bufferSize)
	c.stats = new(ClientStats)
	c.attributes = make(map[string]interface{})
	return c
}
