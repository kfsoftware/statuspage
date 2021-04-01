package check

import (
	"net"
	"time"
)

type TcpCheck struct {
	addr string
}
type TcpStatistics struct {
	TimeTaken  time.Duration
	RemoteAddr string
}

func (i TcpStatistics) GetTimeTaken() time.Duration {
	return i.TimeTaken
}

func (h TcpCheck) GetType() Type {
	return TcpType
}

func (h TcpCheck) Check() Result {
	result := Result{}
	statistics := TcpStatistics{}
	start := time.Now()
	resp, err := net.DialTimeout("tcp", h.addr, 10*time.Second)
	end := time.Now()
	statistics.TimeTaken = end.Sub(start)
	if err != nil {
		result.Statistics = statistics
		result.Error = err
		result.Message = err.Error()
		return result
	}
	defer resp.Close()
	statistics.RemoteAddr = resp.RemoteAddr().String()
	result.Statistics = statistics
	return result
}

func NewTcpCheck(addr string) Check {
	return TcpCheck{
		addr,
	}
}
