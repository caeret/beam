package beam

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadRequest(t *testing.T) {
	assert := assert.New(t)
	var (
		req Request
		b   []byte
		l   []byte
		err error
	)
	b = []byte("*1\r\n$4\r\nPING\r\n")
	req, l, err = ReadRequest(b)
	assert.Nil(err)
	assert.Empty(l)
	assert.Equal(req, Request{[]byte("PING")})

	b = []byte("*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n")
	req, l, err = ReadRequest(b)
	assert.Nil(err)
	assert.Empty(l)
	assert.Equal(Request{[]byte("GET"), []byte("foo")}, req)

	b = []byte("*2\r\nhello")
	req, l, err = ReadRequest(b)
	assert.Equal(ErrFormat, err)
	assert.Empty(l)

	b = []byte("*1\r\n$4\r\nPING\r\n*2\r\n$3GET\r")
	req, l, err = ReadRequest(b)
	assert.Nil(err)
	assert.Equal(Request{[]byte("PING")}, req)
	assert.Equal([]byte("*2\r\n$3GET\r"), l)

	b = []byte("*2\r\n$3GET\r")
	req, l, err = ReadRequest(b)
	assert.Equal(io.EOF, err)
	assert.Equal(b, l)
}

func TestRequest_String(t *testing.T) {
	assert := assert.New(t)
	req1 := Request{[]byte("GET"), []byte("bar")}
	assert.Equal(req1.String(), "*2\\r\\n$3\\r\\nGET\\r\\n$3\\r\\nbar\\r\\n")

	req2 := Request{[]byte("PING")}
	assert.Equal(req2.String(), "*1\\r\\n$4\\r\\nPING\\r\\n")
}

func TestRequest_Raw(t *testing.T) {
	assert := assert.New(t)
	req1 := Request{[]byte("GET"), []byte("bar")}
	assert.Equal(req1.Raw(), "*2\r\n$3\r\nGET\r\n$3\r\nbar\r\n")

	req2 := Request{[]byte("PING")}
	assert.Equal(req2.Raw(), "*1\r\n$4\r\nPING\r\n")
}

func TestNewRequest(t *testing.T) {
	assert := assert.New(t)
	req := NewRequest("PING")
	assert.Len(req, 1)
	assert.Equal([]byte("PING"), req[0])
}
