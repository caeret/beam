package beam

type HandlerFunc func(request Request) (Response, error)

func (hf HandlerFunc) Handle(request Request) (Response, error) {
	return hf(request)
}

// Handler handles the server request.
type Handler interface {
	Handle(request Request) (Response, error)
}
