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

// Commands is a Command list.
type Commands []Command

func (cs Commands) String() string {
	return escapeCrlf(cs.Raw())
}

// Raw formats the Command to redis binary protocol.
func (cs Commands) Raw() string {
	strs := make([]string, len(cs))
	for i, req := range cs {
		strs[i] = req.Raw()
	}
	return strings.Join(strs, "")
}

// NewCommand creates new Command from string arguments.
func NewCommand(args ...string) Command {
	cmd := make(Command, len(args))
	for i, arg := range args {
		cmd[i] = []byte(arg)
	}
	return cmd
}

// Command represents the redis Command command.
type Command [][]byte

// Get retrieves the arg bytes with the given index.
func (cmd Command) Get(index int) []byte {
	if index < len(cmd) {
		return cmd[index]
	}
	return nil
}

// GetStr retrieves the arg string with the given index.
func (cmd Command) GetStr(index int) string {
	return string(cmd.Get(index))
}

func (cmd Command) String() string {
	return escapeCrlf(cmd.Raw())
}

// Len retrieves the length of the command.
func (cmd Command) Len() int {
	return len(cmd)
}

// Raw formats the Command to redis binary protocol.
func (cmd Command) Raw() string {
	raw := "*" + strconv.Itoa(len(cmd)) + string(crlf)
	for _, elem := range cmd {
		raw += "$" + strconv.Itoa(len(elem)) + string(crlf)
		raw += string(elem) + string(crlf)
	}
	return raw
}

// ReadCommand parses a Command from b, and the left bytes l will be returned.
// ErrFormat will be returned if there is invalid protocol sequence.
func ReadCommand(b []byte) (cmd Command, l []byte, err error) {
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

	cmd = make(Command, 0, cnt)

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

		cmd = append(cmd, arg)
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
