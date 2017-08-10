package beam

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
)

var (
	ErrFormat = errors.New("invalid format")
)

// NewRequest creates new request from string arguments.
func NewRequest(args ...string) Request {
	req := make(Request, len(args))
	for i, arg := range args {
		req[i] = []byte(arg)
	}
	return req
}

// Request represents the redis request command.
type Request [][]byte

func (r Request) String() string {
	return escapeCrlf(r.Raw())
}

func (r Request) Raw() string {
	raw := "*" + strconv.Itoa(len(r)) + string(crlf)
	for _, elem := range r {
		raw += "$" + strconv.Itoa(len(elem)) + string(crlf)
		raw += string(elem) + string(crlf)
	}
	return raw
}

// ReadRequest reads a request from the io.Reader.
func ReadRequest(reader io.Reader) (request Request, err error) {
	bReader := bufio.NewReader(reader)
	n, err := readArgsCount(bReader)
	if err != nil {
		return
	}

	request = make(Request, n)

	for i := 0; i < n; i++ {
		var argLength int
		argLength, err = readArgLength(bReader)
		if err != nil {
			return
		}
		var arg []byte
		arg, err = readArgLine(argLength, bReader)
		if err != nil {
			return
		}

		request[i] = arg
	}

	return
}

// readArgsCount reads arguments count from reader.
func readArgsCount(reader *bufio.Reader) (n int, err error) {
	argsCount, err := readPrefixedLine('*', reader)
	if err != nil {
		return
	}
	n, err = strconv.Atoi(string(argsCount))
	return
}

// readArgLength reads the length of argument from reader.
func readArgLength(reader *bufio.Reader) (n int, err error) {
	argLength, err := readPrefixedLine('$', reader)
	if err != nil {
		return
	}
	n, err = strconv.Atoi(string(argLength))
	return
}

// readPrefixedLine reads the line starts by the specified prefix and ends with crlf.
func readPrefixedLine(prefix byte, reader *bufio.Reader) (line []byte, err error) {
	b, err := reader.ReadByte()
	if err != nil {
		return
	}
	if b != prefix {
		err = ErrFormat
		return
	}
	line, err = reader.ReadBytes('\n')
	if err != nil {
		return
	}
	if !bytes.HasSuffix(line, crlf) {
		err = ErrFormat
		return
	}
	line = line[:len(line)-2]
	return
}

// readArgLine reads the line with specified length.
func readArgLine(length int, reader *bufio.Reader) (line []byte, err error) {
	line = make([]byte, length+2)
	var (
		n        int
		currentN int
	)
	for n < length {
		currentN, err = reader.Read(line[n:])
		if err != nil {
			return
		}
		n += currentN
	}
	if !bytes.HasSuffix(line, crlf) {
		err = ErrFormat
		return
	}
	line = line[:length]
	return
}
