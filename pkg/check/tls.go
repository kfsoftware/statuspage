package check

import (
	"crypto/tls"
	"github.com/pkg/errors"
	"time"
)

type TlsCheck struct {
	addr      string
	tlsConfig *tls.Config
}

type PeerCertificate struct {
	Content []byte
}

type TlsStatistics struct {
	TimeTaken        time.Duration
	PeerCertificates []PeerCertificate
}

func (i TlsStatistics) GetTimeTaken() time.Duration {
	return i.TimeTaken
}

func (h TlsCheck) GetType() Type {
	return TlsType
}

func (h TlsCheck) Check() Result {
	result := Result{}
	statistics := TlsStatistics{}
	start := time.Now()
	conn, err := tls.Dial("tcp", h.addr, h.tlsConfig)
	end := time.Now()
	statistics.TimeTaken = end.Sub(start)
	if err != nil {
		result.Statistics = statistics
		result.Error = err
		result.Message = err.Error()
		return result
	}
	defer conn.Close()
	defer conn.CloseWrite()
	err = conn.Handshake()
	if err != nil {
		result.Statistics = statistics
		result.Error = err
		result.Message = err.Error()
		return result
	}
	peerCertificates := conn.ConnectionState().PeerCertificates
	for _, peerCertificate := range peerCertificates {
		statistics.PeerCertificates = append(statistics.PeerCertificates, PeerCertificate{Content: peerCertificate.Raw})
	}

	expiry := peerCertificates[0].NotAfter
	if expiry.Before(time.Now()) {
		err = errors.Errorf("Certificate expired")
		result.Error = err
		result.Statistics = statistics
		result.Message = err.Error()
		return result
	}
	result.Statistics = statistics
	return result
}

func NewTlsCheck(url string, tlsConfig *tls.Config) Check {
	return TlsCheck{
		url,
		tlsConfig,
	}
}
