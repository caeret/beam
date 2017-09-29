package beam

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadInlineQuery(t *testing.T) {
	assert := assert.New(t)
	var (
		querys []Query
		b      []byte
		l      []byte
		err    error
	)

	b = []byte("GET foo\r\nGET bar\n")
	querys, l, err = ReadQuery(b)
	assert.Nil(err)
	assert.Equal(
		[]Query{
			NewQuery("GET", "foo"),
			NewQuery("GET", "bar"),
		}, querys)
	assert.Empty(l)

	b = []byte("GET foo\r\nGET")
	querys, l, err = ReadQuery(b)
	assert.Nil(err)
	assert.Equal(
		[]Query{
			NewQuery("GET", "foo"),
		}, querys)
	assert.Equal([]byte("GET"), l)
}

func TestReadMultiBuckQuery(t *testing.T) {
	assert := assert.New(t)
	var (
		querys []Query
		b      []byte
		l      []byte
		err    error
	)

	b = []byte("*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n*2\r\n$3\r\nGET\r\n$3\r\nbar\r\n")
	querys, l, err = ReadQuery(b)
	assert.Nil(err)
	assert.Empty(l)
	assert.Equal(
		[]Query{
			NewQuery("GET", "foo"),
			NewQuery("GET", "bar"),
		},
		querys)

	b = []byte("*2\r\nhello")
	querys, l, err = ReadQuery(b)
	assert.Equal(ErrFormat, err)
	assert.Empty(l)

	b = []byte("*1\r\n$4\r\nPING\r\n*2\r\n$3GET\r")
	querys, l, err = ReadQuery(b)
	assert.Nil(err)
	assert.Equal([]Query{NewQuery("PING")}, querys)
	assert.Equal([]byte("*2\r\n$3GET\r"), l)

	b = []byte("*2\r\n$3GET\r")
	querys, l, err = ReadQuery(b)
	assert.Nil(err)
	assert.Equal(b, l)
}

func TestQuery_String(t *testing.T) {
	assert := assert.New(t)
	req1 := NewQuery("GET", "bar")
	assert.Equal(req1.String(), "*2\\r\\n$3\\r\\nGET\\r\\n$3\\r\\nbar\\r\\n")

	req2 := NewQuery("PING")
	assert.Equal(req2.String(), "*1\\r\\n$4\\r\\nPING\\r\\n")
}

func TestQuery_Raw(t *testing.T) {
	assert := assert.New(t)
	req1 := NewQuery("GET", "bar")
	assert.Equal(req1.Raw(), "*2\r\n$3\r\nGET\r\n$3\r\nbar\r\n")

	req2 := NewQuery("PING")
	assert.Equal(req2.Raw(), "*1\r\n$4\r\nPING\r\n")
}

func TestNewQuery(t *testing.T) {
	assert := assert.New(t)
	query := NewQuery("PING")
	assert.Equal(query.Len(), 0)
	assert.Equal([]byte("PING"), query.Command)
}
