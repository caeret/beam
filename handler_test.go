package beam

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testHandlerChain(t *testing.T) {
	assert := assert.New(t)
	chain := NewHandlerChain(HandleFunc(func(request *Request) (Reply, error) {
		return NewSimpleStringsReply("FOO"), nil
	}))
	req := &Request{
		Command: Command{[]byte("GET")},
	}
	reply, err := chain.Handle(req)
	assert.Nil(err)
	assert.EqualValues([]byte("+FOO"), reply)

	chain.AddFunc(func(request *Request, next Handler) (Reply, error) {
		if request.GetStr(0) == "FOO" {
			return next.Handle(req)
		}
		return NewSimpleStringsReply("BAR"), nil
	})

	reply, err = chain.Handle(req)
	assert.Nil(err)
	assert.EqualValues([]byte("+FOO"), reply)

	req = &Request{
		Command: Command{[]byte("FOO")},
	}
	reply, err = chain.Handle(req)
	assert.Nil(err)
	assert.EqualValues([]byte("+BAR"), reply)
}
