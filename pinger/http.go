package pinger

import (
	"net/http"
	"time"
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

func HTTPCheck(cfg HTTPCheckConfig) HTTPResult {
	start := time.Now()
	client := &http.Client{Timeout: cfg.Timeout}
	resp, err := client.Get(cfg.URL)
	if err != nil {
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
