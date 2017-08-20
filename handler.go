package beam

type HandlerFunc func(request *Request) (Reply, error)

func (hf HandlerFunc) Handle(request *Request) (Reply, error) {
	return hf(request)
}

// Handler handles the server request.
type Handler interface {
	Handle(request *Request) (Reply, error)
}
