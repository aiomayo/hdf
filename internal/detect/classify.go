package detect

import (
	"net"
	"strconv"
	"strings"
)

type QueryType int

const (
	TypePort QueryType = iota
	TypePID
	TypeHostPort
	TypeGlob
	TypeName
)

func (t QueryType) String() string {
	switch t {
	case TypePort:
		return "port"
	case TypePID:
		return "pid"
	case TypeHostPort:
		return "host:port"
	case TypeGlob:
		return "glob"
	case TypeName:
		return "name"
	default:
		return "unknown"
	}
}

type Query struct {
	Type QueryType
	Raw  string
	Port uint32
	PID  int32
	Name string
}

func Classify(input string) Query {
	input = strings.TrimSpace(input)
	q := Query{Raw: input}

	if num, err := strconv.ParseUint(input, 10, 64); err == nil {
		if num >= 1 && num <= 65535 {
			q.Type = TypePort
			q.Port = uint32(num)
			return q
		}
		q.Type = TypePID
		q.PID = int32(num)
		return q
	}

	if strings.Contains(input, ":") {
		_, portStr, err := net.SplitHostPort(input)
		if err == nil {
			if port, err := strconv.ParseUint(portStr, 10, 32); err == nil && port >= 1 && port <= 65535 {
				q.Type = TypeHostPort
				q.Port = uint32(port)
				return q
			}
		}
	}

	if strings.ContainsAny(input, "*?") {
		q.Type = TypeGlob
		q.Name = input
		return q
	}

	q.Type = TypeName
	q.Name = input
	return q
}
