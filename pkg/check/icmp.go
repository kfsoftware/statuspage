package check

import (
	"github.com/go-ping/ping"
	"time"
)

type IcmpCheck struct {
	addr string
}
type IcmpStatistics struct {
	TimeTaken      time.Duration
	PingStatistics *ping.Statistics
}

func (i IcmpStatistics) GetTimeTaken() time.Duration {
	return i.TimeTaken
}

func (h IcmpCheck) GetType() Type {
	return IcmpType
}

func (h IcmpCheck) Check() (result Result) {
	statistics := IcmpStatistics{}
	result.Statistics = statistics
	start := time.Now()
	pinger, err := ping.NewPinger(h.addr)
	if err != nil {
		result.Error = err
		result.Message = err.Error()
		return
	}
	pinger.Count = 1
	err = pinger.Run()
	end := time.Now()
	statistics.TimeTaken = end.Sub(start)
	if err != nil {
		result.Error = err
		result.Message = err.Error()
		return
	}
	stats := pinger.Statistics() // get send/receive/duplicate/rtt stats
	statistics.PingStatistics = stats
	return
}

func NewIcmpCheck(addr string) Check {
	return IcmpCheck{
		addr,
	}
}
