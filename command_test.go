package beam

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadCommand(t *testing.T) {
	assert := assert.New(t)
	var (
		cmds []Command
		b    []byte
		l    []byte
		err  error
	)
	b = []byte("*1\r\n$4\r\nPING\r\n")
	cmds, l, err = ReadCommand(b)
	assert.Nil(err)
	assert.Empty(l)
	assert.Equal([]Command{{[]byte("PING")}}, cmds)

	b = []byte("*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n")
	cmds, l, err = ReadCommand(b)
	assert.Nil(err)
	assert.Empty(l)
	assert.Equal([]Command{{[]byte("GET"), []byte("foo")}}, cmds)

	b = []byte("*2\r\nhello")
	cmds, l, err = ReadCommand(b)
	assert.Equal(ErrFormat, err)
	assert.Nil(l)

	b = []byte("*1\r\n$4\r\nPING\r\n*2\r\n$3GET\r")
	cmds, l, err = ReadCommand(b)
	assert.Nil(err)
	assert.Equal([]Command{{[]byte("PING")}}, cmds)
	assert.Equal([]byte("*2\r\n$3GET\r"), l)

	b = []byte("*2\r\n$3GET\r")
	cmds, l, err = ReadCommand(b)
	assert.Nil(err)
	assert.Equal(b, l)
}

func TestCommand_String(t *testing.T) {
	assert := assert.New(t)
	req1 := Command{[]byte("GET"), []byte("bar")}
	assert.Equal(req1.String(), "*2\\r\\n$3\\r\\nGET\\r\\n$3\\r\\nbar\\r\\n")

	req2 := Command{[]byte("PING")}
	assert.Equal(req2.String(), "*1\\r\\n$4\\r\\nPING\\r\\n")
}

func TestCommand_Raw(t *testing.T) {
	assert := assert.New(t)
	req1 := Command{[]byte("GET"), []byte("bar")}
	assert.Equal(req1.Raw(), "*2\r\n$3\r\nGET\r\n$3\r\nbar\r\n")

	req2 := Command{[]byte("PING")}
	assert.Equal(req2.Raw(), "*1\r\n$4\r\nPING\r\n")
}

func TestNewCommand(t *testing.T) {
	assert := assert.New(t)
	req := NewCommand("PING")
	assert.Len(req, 1)
	assert.Equal([]byte("PING"), req[0])
}
