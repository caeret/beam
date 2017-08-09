package beam

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSimpleStringsResponse(t *testing.T) {
	assert := assert.New(t)

	var resp Response
	resp = NewSimpleStringsResponse("OK")
	assert.Equal(Response("+OK\r\n"), resp)
}

func TestNewErrorsResponse(t *testing.T) {
	assert := assert.New(t)

	var resp Response
	resp = NewErrorsResponse("WRONGTYPE")
	assert.Equal(Response("-WRONGTYPE\r\n"), resp)
}

func TestNewIntegersResponse(t *testing.T) {
	assert := assert.New(t)

	var resp Response
	resp = NewIntegersResponse(10)
	assert.Equal(Response(":10\r\n"), resp)
}

func TestNewBulkStringsResponse(t *testing.T) {
	assert := assert.New(t)

	var resp Response
	resp = NewBulkStringsResponse("foo")
	assert.Equal(Response("$3\r\nfoo\r\n"), resp)
}

func TestNewBulkStringsResponseRaw(t *testing.T) {
	assert := assert.New(t)

	var resp Response
	resp = NewBulkStringsResponseRaw(nil)
	assert.Equal(Response("$-1\r\n"), resp)
}

func TestNewArraysResponse(t *testing.T) {
	assert := assert.New(t)

	var resp Response
	resp = NewArraysResponse("foo", "bar")
	assert.Equal(Response("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"), resp)
}

func TestNewArraysResponseRaw(t *testing.T) {
	assert := assert.New(t)

	var resp Response
	resp = NewArraysResponseRaw([]byte("foo"), nil)
	assert.Equal(Response("*2\r\n$3\r\nfoo\r\n$-1\r\n"), resp)
}
