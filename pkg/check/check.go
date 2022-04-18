package check

import "time"

type Statistics interface {
	GetTimeTaken() time.Duration
}

type Result struct {
	Error      error
	Message    string
	Warnings   []Warning
	Statistics Statistics
}
type Warning struct {
	Message string
}
type Type string

const (
	HttpType Type = "http"
	IcmpType Type = "icmp"
	TcpType  Type = "tcp"
	TlsType  Type = "tls"
)

type Check interface {
	GetType() Type
	Check() Result
}
