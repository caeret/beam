package beam

type HandlerFunc func(request *ClientRequest) (Response, error)

func (hf HandlerFunc) Handle(request *ClientRequest) (Response, error) {
	return hf(request)
}

// Handler handles the server request.
type Handler interface {
	Handle(request *ClientRequest) (Response, error)
}
