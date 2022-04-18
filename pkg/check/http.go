package check

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

type HttpCheck struct {
	url                string
	certificateWarning time.Duration
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

func (h HttpCheck) Check() Result {
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
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
	if resp.TLS != nil && resp.TLS.VerifiedChains != nil && len(resp.TLS.VerifiedChains) > 0 {
		if time.Now().After(resp.TLS.VerifiedChains[0][0].NotAfter.Add(-h.certificateWarning)) {
			result.Warnings = []Warning{
				{
					Message: fmt.Sprintf("Certificate for %s is about to expire in %s", h.url, humanizeDuration(resp.TLS.VerifiedChains[0][0].NotAfter.Sub(time.Now()))),
				},
			}
		}
	}
	return result
}

// humanizeDuration humanizes time.Duration output to a meaningful value,
// golang's default ``time.Duration`` output is badly formatted and unreadable.
func humanizeDuration(duration time.Duration) string {
	if duration.Seconds() < 60.0 {
		return fmt.Sprintf("%d seconds", int64(duration.Seconds()))
	}
	if duration.Minutes() < 60.0 {
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%d minutes %d seconds", int64(duration.Minutes()), int64(remainingSeconds))
	}
	if duration.Hours() < 24.0 {
		remainingMinutes := math.Mod(duration.Minutes(), 60)
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%d hours %d minutes %d seconds",
			int64(duration.Hours()), int64(remainingMinutes), int64(remainingSeconds))
	}
	remainingHours := math.Mod(duration.Hours(), 24)
	remainingMinutes := math.Mod(duration.Minutes(), 60)
	remainingSeconds := math.Mod(duration.Seconds(), 60)
	return fmt.Sprintf("%d days %d hours %d minutes %d seconds",
		int64(duration.Hours()/24), int64(remainingHours),
		int64(remainingMinutes), int64(remainingSeconds))
}
func NewHttpCheck(url string, expectedStatusCode *int, certificateWarningDelta time.Duration) Check {
	return HttpCheck{
		certificateWarning: certificateWarningDelta, // time.Hour * 24 * 15,
		url:                url,
		expectedStatusCode: expectedStatusCode,
	}
}
