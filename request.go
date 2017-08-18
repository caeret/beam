package beam

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
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

type Requests []Request

func (rs Requests) String() string {
	return escapeCrlf(rs.Raw())
}

func (rs Requests) Raw() string {
	strs := make([]string, len(rs))
	for i, req := range rs {
		strs[i] = req.Raw()
	}
	return strings.Join(strs, "")
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

func ReadRequest(b []byte) (request Request, l []byte, err error) {
	defer func() {
		if err != nil && err == io.EOF {
			l = b
		}
	}()
	reader := bytes.NewBuffer(b)
	cnt, err := readArgsCount(reader)
	if err != nil {
		return
	}

	request = make(Request, 0, cnt)

	for i := 0; i < cnt; i++ {
		var argLength int
		argLength, err = readArgLength(reader)
		if err != nil {
			return
		}
		var arg []byte
		arg, err = readArgLine(argLength, reader)
		if err != nil {
			return
		}

		request = append(request, arg)
	}

	buff := new(bytes.Buffer)
	_, err = reader.WriteTo(buff)
	if err != nil {
		return
	}
	l = buff.Bytes()
	return
}

// readArgsCount reads arguments count from reader.
func readArgsCount(reader *bytes.Buffer) (n int, err error) {
	argsCount, err := readPrefixedLine('*', reader)
	if err != nil {
		return
	}
	n, err = strconv.Atoi(string(argsCount))
	return
}

// readArgLength reads the length of argument from reader.
func readArgLength(reader *bytes.Buffer) (n int, err error) {
	argLength, err := readPrefixedLine('$', reader)
	if err != nil {
		return
	}
	n, err = strconv.Atoi(string(argLength))
	return
}

// readPrefixedLine reads the line starts by the specified prefix and ends with crlf.
func readPrefixedLine(prefix byte, reader *bytes.Buffer) (line []byte, err error) {
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
func readArgLine(length int, reader *bytes.Buffer) (line []byte, err error) {
	line = make([]byte, length+2)
	n, err := reader.Read(line)
	if err != nil {
		return
	}

	if n != len(line) {
		return nil, io.EOF
	}

	if !bytes.HasSuffix(line, crlf) {
		err = ErrFormat
		return
	}
	line = line[:length]
	return
}
