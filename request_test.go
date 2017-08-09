package beam

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadRequest(t *testing.T) {
	assert := assert.New(t)
	var (
		reader io.Reader
		req    Request
		err    error
	)
	reader = strings.NewReader("*1\r\n$4\r\nPING\r\n")
	req, err = ReadRequest(reader)
	assert.Nil(err)
	assert.Equal(req, Request{[]byte("PING")})

	reader = strings.NewReader("*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n")
	req, err = ReadRequest(reader)
	assert.Nil(err)
	assert.Equal(req, Request{[]byte("GET"), []byte("foo")})

	reader = strings.NewReader("*2\r\nhello")
	req, err = ReadRequest(reader)
	assert.Error(ErrFormat, err)
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
