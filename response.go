package beam

import (
	"bytes"
	"strconv"
)

const (
	SimpleStringsResponsePrefix = '+'
	ErrorsResponsePrefix        = '-'
	IntegersResponsePrefix      = ':'
	BulkStringsResponsePrefix   = '$'
	ArraysResponsePrefix        = '*'
)

// Response represent the redis response.
type Response []byte

func (r Response) String() string {
	return escapeCrlf(r.Raw())
}

func (r Response) Raw() string {
	return string(r)
}

// NewSimpleStringsResponse create the response for simple strings reply.
func NewSimpleStringsResponse(data string) Response {
	return createSimpleResponse(SimpleStringsResponsePrefix, data)
}

// NewErrorsResponse create the response for errors reply.
func NewErrorsResponse(data string) Response {
	return createSimpleResponse(ErrorsResponsePrefix, data)
}

// NewIntegersResponse create the response for integers reply.
func NewIntegersResponse(data int) Response {
	return createSimpleResponse(IntegersResponsePrefix, strconv.Itoa(data))
}

// NewBulkStringsResponse create the response for binary-safe strings reply.
func NewBulkStringsResponse(data string) Response {
	return createBuckStrings([]byte(data))
}

// NewBulkStringsResponseRaw create the response for binary-safe strings reply with raw bytes.
func NewBulkStringsResponseRaw(data []byte) Response {
	return createBuckStrings(data)
}

// NewArraysResponse create the response for arrays reply.
func NewArraysResponse(data ...string) Response {
	buffer := make([][]byte, 3+len(data))
	buffer[0] = []byte{ArraysResponsePrefix}
	buffer[1] = []byte(strconv.Itoa(len(data)))
	buffer[2] = crlf
	for i, elem := range data {
		buffer[3+i] = createBuckStrings([]byte(elem))
	}
	return bytes.Join(buffer, nil)
}

// NewArraysResponseRaw create the response for arrays reply with raw bytes.
func NewArraysResponseRaw(data ...[]byte) Response {
	buffer := make([][]byte, 3+len(data))
	buffer[0] = []byte{ArraysResponsePrefix}
	buffer[1] = []byte(strconv.Itoa(len(data)))
	buffer[2] = crlf
	for i, elem := range data {
		buffer[3+i] = createBuckStrings(elem)
	}
	return bytes.Join(buffer, nil)
}

func createSimpleResponse(prefix byte, data string) []byte {
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
		buffer[0] = []byte{BulkStringsResponsePrefix}
		buffer[1] = []byte("-1")
		buffer[2] = crlf
	} else {
		buffer = make([][]byte, 5)
		buffer[0] = []byte{BulkStringsResponsePrefix}
		buffer[1] = []byte(strconv.Itoa(len(data)))
		buffer[2] = crlf
		buffer[3] = []byte(data)
		buffer[4] = crlf
	}
	return bytes.Join(buffer, nil)
}
