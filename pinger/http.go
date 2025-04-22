package pinger

import (
	"crypto/x509"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)



type HTTPCheckConfig struct {
	URL              string
	Timeout          time.Duration
	AcceptStatusCodes []int
}

type HTTPResult struct {
	Up           bool
	RespTime     float64
	StatusCode   int
	SSLDaysLeft  int
	Err          error
}

func HTTPCheck(cfg HTTPCheckConfig, sslErrorCounter *prometheus.CounterVec) HTTPResult {
	start := time.Now()
	client := &http.Client{Timeout: cfg.Timeout}
	resp, err := client.Get(cfg.URL)
	if err != nil {
		// SSL error detection
		var errType string
		if _, ok := err.(x509.UnknownAuthorityError); ok {
			errType = "UnknownAuthority"
		} else if _, ok := err.(x509.CertificateInvalidError); ok {
			errType = "CertificateInvalid"
		} else if err != nil && err.Error() != "" && (containsTLS(err.Error())) {
			errType = "TLS"
		} else {
			errType = "Unknown"
		}
		if sslErrorCounter != nil {
			sslErrorCounter.WithLabelValues(cfg.URL, errType).Inc()
		}
		return HTTPResult{Up: false, RespTime: 0, StatusCode: 0, SSLDaysLeft: -1, Err: err}
	}
	defer resp.Body.Close()

	sslDays := -1
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		expiry := resp.TLS.PeerCertificates[0].NotAfter
		sslDays = int(time.Until(expiry).Hours() / 24)
	}

	accepted := false
	for _, code := range cfg.AcceptStatusCodes {
		if resp.StatusCode == code {
			accepted = true
			break
		}
	}

	return HTTPResult{
		Up:           accepted,
		RespTime:     time.Since(start).Seconds(),
		StatusCode:   resp.StatusCode,
		SSLDaysLeft:  sslDays,
		Err:          nil,
	}
}

// containsTLS checks if the error string contains 'tls:'
func containsTLS(s string) bool {
	return len(s) > 0 && (len(s) >= 4 && s[:4] == "tls:")
}
