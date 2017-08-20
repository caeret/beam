package beam

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSimpleStringsReply(t *testing.T) {
	assert := assert.New(t)

	var resp Reply
	resp = NewSimpleStringsReply("OK")
	assert.Equal(Reply("+OK\r\n"), resp)
}

func TestNewErrorsReply(t *testing.T) {
	assert := assert.New(t)

	var resp Reply
	resp = NewErrorsReply("WRONGTYPE")
	assert.Equal(Reply("-WRONGTYPE\r\n"), resp)
}

func TestNewIntegersReply(t *testing.T) {
	assert := assert.New(t)

	var resp Reply
	resp = NewIntegersReply(10)
	assert.Equal(Reply(":10\r\n"), resp)
}

func TestNewBulkStringsReply(t *testing.T) {
	assert := assert.New(t)

	var resp Reply
	resp = NewBulkStringsReply("foo")
	assert.Equal(Reply("$3\r\nfoo\r\n"), resp)
}

func TestNewBulkStringsReplyRaw(t *testing.T) {
	assert := assert.New(t)

	var resp Reply
	resp = NewBulkStringsReplyRaw(nil)
	assert.Equal(Reply("$-1\r\n"), resp)
}

func TestNewArraysReply(t *testing.T) {
	assert := assert.New(t)

	var resp Reply
	resp = NewArraysReply("foo", "bar")
	assert.Equal(Reply("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"), resp)
}

func TestNewArraysReplyRaw(t *testing.T) {
	assert := assert.New(t)

	var resp Reply
	resp = NewArraysReplyRaw([]byte("foo"), nil)
	assert.Equal(Reply("*2\r\n$3\r\nfoo\r\n$-1\r\n"), resp)
}
