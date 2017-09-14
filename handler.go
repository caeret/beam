package beam

import (
	"container/list"
	"errors"
	"strings"
)

var (
	ErrHandlerNotFound = errors.New("handler not found")
)

type HandleFunc func(request *Request) (Reply, error)

func (hf HandleFunc) Handle(request *Request) (Reply, error) {
	return hf(request)
}

// Handler handles the server request.
type Handler interface {
	Handle(request *Request) (Reply, error)
}

type Middleware interface {
	Do(request *Request, next Handler) (Reply, error)
}

type MiddlewareFunc func(request *Request, next Handler) (Reply, error)

func (m MiddlewareFunc) Do(request *Request, next Handler) (Reply, error) {
	return m(request, next)
}

// NewHandlerChain creates a new chained handlers.
func NewHandlerChain(h Handler) *HandlerChain {
	hc := new(HandlerChain)
	hc.handler = h
	hc.list = list.New()
	hc.list.PushFront(h)
	return hc
}

type HandlerChain struct {
	handler Handler
	list    *list.List
}

func (hc *HandlerChain) Add(m Middleware) {
	next := hc.handler
	hc.handler = HandleFunc(func(request *Request) (Reply, error) {
		return m.Do(request, next)
	})
	hc.list.PushFront(m)
}

func (hc *HandlerChain) AddFunc(f func(request *Request, next Handler) (Reply, error)) {
	hc.Add(MiddlewareFunc(f))
}

func (hc *HandlerChain) Handle(request *Request) (Reply, error) {
	return hc.handler.Handle(request)
}

func NewMappedHandler() *MappedHandler {
	mh := new(MappedHandler)
	mh.handlers = make(map[string]Handler)
	return mh
}

type MappedHandler struct {
	handlers map[string]Handler
}

func (mh *MappedHandler) Set(command string, h Handler) {
	mh.handlers[strings.ToUpper(command)] = h
}

func (mh *MappedHandler) SetFunc(command string, f func(request *Request) (Reply, error)) {
	mh.Set(command, HandleFunc(f))
}

func (mh *MappedHandler) Handle(request *Request) (Reply, error) {
	command := strings.ToUpper(request.GetStr(0))
	if handler, exist := mh.handlers[command]; exist {
		return handler.Handle(request)
	}
	return nil, ErrHandlerNotFound
}
