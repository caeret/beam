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

// Querys is a Query list.
type Querys []Query

func (qs Querys) String() string {
	return escapeCrlf(qs.Raw())
}

// Raw formats the Query to redis binary protocol.
func (qs Querys) Raw() string {
	strs := make([]string, len(qs))
	for i, req := range qs {
		strs[i] = req.Raw()
	}
	return strings.Join(strs, "")
}

// NewQuery creates new Query from string arguments.
func NewQuery(directive string, args ...string) Query {
	var query Query
	query.Command = []byte(directive)
	query.Arguments = make([][]byte, len(args))
	for i, arg := range args {
		query.Arguments[i] = []byte(arg)
	}
	return query
}

// Query represents the redis Query query.
type Query struct {
	Command   []byte
	Arguments [][]byte
}

// CommandStr return the string type of command.
func (query Query) CommandStr() string {
	return string(query.Command)
}

// Arg retrieves the arg bytes with the given index.
func (query Query) Arg(index int) []byte {
	if index < len(query.Arguments) {
		return query.Arguments[index]
	}
	return nil
}

// ArgStr retrieves the arg string with the given index.
func (query Query) ArgStr(index int) string {
	return string(query.Arg(index))
}

func (query Query) String() string {
	return escapeCrlf(query.Raw())
}

// Len retrieves the arguments length of the query.
func (query Query) Len() int {
	return len(query.Arguments)
}

// Raw formats the Query to redis binary protocol.
func (query Query) Raw() string {
	raw := "*" + strconv.Itoa(query.Len()+1) + string(crlf)
	raw += "$" + strconv.Itoa(len(query.Command)) + string(crlf)
	raw += string(query.Command) + string(crlf)
	for _, elem := range query.Arguments {
		raw += "$" + strconv.Itoa(len(elem)) + string(crlf)
		raw += string(elem) + string(crlf)
	}
	return raw
}

// ReadQuery parses querys from b, and the left bytes l will be returned.
// ErrFormat will be returned if there is invalid protocol sequence.
func ReadQuery(b []byte) (querys []Query, l []byte, err error) {
	buffer := bytes.NewBuffer(b)
	// assumes that each query needs 32 bytes.
	querys = make([]Query, 0, len(b)/32)

	defer func() {
		if err == io.EOF {
			err = nil
		} else {
			l = nil
		}
	}()

	for err == nil {
		l = buffer.Bytes()
		if len(l) == 0 {
			err = io.EOF
			break
		}

		var (
			query Query
			empty bool
		)

		if l[0] == '*' {
			query, err = readMultiBulkQuery(buffer)
		} else {
			query, empty, err = readInlineQuery(buffer)
		}

		if err != nil {
			return
		}

		if empty {
			continue
		}

		querys = append(querys, query)
	}

	return
}

func readInlineQuery(buffer *bytes.Buffer) (query Query, empty bool, err error) {
	b, err := buffer.ReadBytes('\n')
	if err != nil {
		return
	}

	args := bytes.Fields(b)

	if len(args) > 0 {
		query.Command = args[0]
		for _, arg := range args[1:] {
			query.Arguments = append(query.Arguments, arg)
		}
	} else {
		empty = true
	}

	return
}

func readMultiBulkQuery(buffer *bytes.Buffer) (query Query, err error) {
	var cnt int
	cnt, err = readArgsCount(buffer)
	if err != nil {
		return
	}

	query.Arguments = make([][]byte, 0, cnt-1)

	for i := 0; i < cnt; i++ {
		var argLength int
		argLength, err = readArgLength(buffer)
		if err != nil {
			return
		}
		var arg []byte
		arg, err = readArgLine(argLength, buffer)
		if err != nil {
			return
		}

		if i == 0 {
			query.Command = arg
		} else {
			query.Arguments = append(query.Arguments, arg)
		}
	}

	return
}

// readArgsCount reads arguments count from buffer.
func readArgsCount(buffer *bytes.Buffer) (n int, err error) {
	argsCount, err := readPrefixedLine('*', buffer)
	if err != nil {
		return
	}
	n, err = strconv.Atoi(string(argsCount))
	return
}

// readArgLength reads the length of argument from buffer.
func readArgLength(buffer *bytes.Buffer) (n int, err error) {
	argLength, err := readPrefixedLine('$', buffer)
	if err != nil {
		return
	}
	n, err = strconv.Atoi(string(argLength))
	return
}

// readPrefixedLine reads the line starts by the specified prefix and ends with crlf.
func readPrefixedLine(prefix byte, buffer *bytes.Buffer) (line []byte, err error) {
	b, err := buffer.ReadByte()
	if err != nil {
		return
	}
	if b != prefix {
		err = ErrFormat
		return
	}
	line, err = buffer.ReadBytes('\n')
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
func readArgLine(length int, buffer *bytes.Buffer) (line []byte, err error) {
	line = make([]byte, length+2)
	n, err := buffer.Read(line)
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
