package checker

import (
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"
)

type PingCheckConfig struct {
	Host    string
	Timeout time.Duration
}

type PingResult struct {
	Up       bool
	RespTime float64
	Err      error
}

func PingCheck(cfg PingCheckConfig) PingResult {
	start := time.Now()
	cmd := exec.Command("ping", "-c", "1", "-W", "2", cfg.Host)
	output, err := cmd.CombinedOutput()
	respTime := time.Since(start).Seconds()
	if os.Getenv("DEBUG_PING_OUTPUT") == "1" {
		log.Printf("[DEBUG] Ping output for %s:\n%s", cfg.Host, string(output))
	}
	if err != nil {
		return PingResult{Up: false, RespTime: respTime, Err: err}
	}
	// Match '0% packet loss' or '0.0% packet loss' as a whole word
	matched, _ := regexp.MatchString(`\b0(\.0)?% packet loss\b`, string(output))
	if matched {
		return PingResult{Up: true, RespTime: respTime, Err: nil}
	}
	return PingResult{Up: false, RespTime: respTime, Err: nil}
}
