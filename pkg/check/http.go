package check

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"time"
)

type HttpCheck struct {
	url                string
	expectedStatusCode *int
}
type HttpStatistics struct {
	TimeTaken     time.Duration
	StatusCode    int
	Headers       map[string][]string
	ContentLength int64
}

func (i HttpStatistics) GetTimeTaken() time.Duration {
	return i.TimeTaken
}
func (h HttpCheck) GetType() Type {
	return HttpType
}

func (h HttpCheck) Check()  Result {
	result := Result{}
	statistics := HttpStatistics{}
	start := time.Now()
	resp, err := http.Get(h.url)
	end := time.Now()
	statistics.TimeTaken = end.Sub(start)
	if err != nil {
		result.Statistics = statistics
		result.Error = err
		result.Message = err.Error()
		return result
	}
	statistics.StatusCode = resp.StatusCode
	statistics.ContentLength = resp.ContentLength
	statistics.Headers = map[string][]string{}
	for k, v := range resp.Header {
		statistics.Headers[strings.ToLower(k)] = v
	}
	if h.expectedStatusCode != nil && resp.StatusCode != *h.expectedStatusCode {
		err = errors.New(fmt.Sprintf("Mismatch status code, expected: %d got: %d", *h.expectedStatusCode, resp.StatusCode))
		result.Statistics = statistics
		result.Error = err
		result.Message = err.Error()
		return result
	}
	result.Message = fmt.Sprintf("Status code: %d", resp.StatusCode)
	result.Statistics = statistics
	return result
}

func NewHttpCheck(url string, expectedStatusCode *int) Check {
	return HttpCheck{
		url,
		expectedStatusCode,
	}
}
