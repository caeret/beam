package beam

import (
	"bytes"
	"strconv"
	"strings"
)

const (
	SimpleStringsReplyPrefix = '+'
	ErrorsReplyPrefix        = '-'
	IntegersReplyPrefix      = ':'
	BulkStringsReplyPrefix   = '$'
	ArraysReplyPrefix        = '*'
)

// Replies is a Reply list.
type Replies []Reply

func (rs Replies) String() string {
	return escapeCrlf(rs.Raw())
}

// Raw formats the reply to redis binary protocol.
func (rs Replies) Raw() string {
	strs := make([]string, len(rs))
	for i, resp := range rs {
		strs[i] = resp.Raw()
	}
	return strings.Join(strs, "")
}

// Reply represents the redis reply.
type Reply []byte

func (r Reply) String() string {
	return escapeCrlf(r.Raw())
}

// Raw formats the reply to redis binary protocol.
func (r Reply) Raw() string {
	return string(r)
}

// NewSimpleStringsReply creates the response for simple strings reply.
func NewSimpleStringsReply(data string) Reply {
	return createSimpleReply(SimpleStringsReplyPrefix, data)
}

// NewErrorsReply creates the response for errors reply.
func NewErrorsReply(data string) Reply {
	return createSimpleReply(ErrorsReplyPrefix, data)
}

// NewIntegersReply creates the response for integers reply.
func NewIntegersReply(data int) Reply {
	return createSimpleReply(IntegersReplyPrefix, strconv.Itoa(data))
}

// NewBulkStringsReply creates the response for binary-safe strings reply.
func NewBulkStringsReply(data string) Reply {
	return createBuckStrings([]byte(data))
}

// NewBulkStringsReplyRaw creates the response for binary-safe strings reply with raw bytes, if the data is nil, the response replies "$-1\r\n".
func NewBulkStringsReplyRaw(data []byte) Reply {
	return createBuckStrings(data)
}

// NewArraysReply creates the response for arrays reply.
func NewArraysReply(data ...string) Reply {
	buffer := make([][]byte, 3+len(data))
	buffer[0] = []byte{ArraysReplyPrefix}
	buffer[1] = []byte(strconv.Itoa(len(data)))
	buffer[2] = crlf
	for i, elem := range data {
		buffer[3+i] = createBuckStrings([]byte(elem))
	}
	return bytes.Join(buffer, nil)
}

// NewArraysReplyRaw creates the response for arrays reply with raw bytes, if the bytes in data is nil, the bulk string in the response replies "$-1\r\n".
func NewArraysReplyRaw(data ...[]byte) Reply {
	buffer := make([][]byte, 3+len(data))
	buffer[0] = []byte{ArraysReplyPrefix}
	buffer[1] = []byte(strconv.Itoa(len(data)))
	buffer[2] = crlf
	for i, elem := range data {
		buffer[3+i] = createBuckStrings(elem)
	}
	return bytes.Join(buffer, nil)
}

func createSimpleReply(prefix byte, data string) []byte {
	buffer := make([][]byte, 3)
	buffer[0] = []byte{prefix}
	buffer[1] = []byte(data)
	buffer[2] = crlf
	return bytes.Join(buffer, nil)
}

func createBuckStrings(data []byte) []byte {
	var buffer [][]byte
	if data == nil {
		buffer = make([][]byte, 3)
		buffer[0] = []byte{BulkStringsReplyPrefix}
		buffer[1] = []byte("-1")
		buffer[2] = crlf
	} else {
		buffer = make([][]byte, 5)
		buffer[0] = []byte{BulkStringsReplyPrefix}
		buffer[1] = []byte(strconv.Itoa(len(data)))
		buffer[2] = crlf
		buffer[3] = []byte(data)
		buffer[4] = crlf
	}
	return bytes.Join(buffer, nil)
}
