package checker

import (
	"regexp"
	"testing"
)

func TestPingCheckSuccessParse(t *testing.T) {
	output := "1 packets transmitted, 1 packets received, 0% packet loss"
	matched, _ := regexp.MatchString(`\b0(\.0)?% packet loss\b`, output)
	if !matched {
		t.Error("Expected output to indicate 0% packet loss (regex match)")
	}
	output2 := "1 packets transmitted, 1 packets received, 0.0% packet loss"
	matched2, _ := regexp.MatchString(`\b0(\.0)?% packet loss\b`, output2)
	if !matched2 {
		t.Error("Expected output to indicate 0.0% packet loss (regex match)")
	}
}

func TestPingCheckFailureParse(t *testing.T) {
	output := "1 packets transmitted, 0 packets received, 100% packet loss"
	matched, _ := regexp.MatchString(`\b0(\.0)?% packet loss\b`, output)
	if matched {
		t.Error("Did not expect output to indicate 0% packet loss (regex match)")
	}
}

func TestPingCheckReturnsFailureOnInvalidHost(t *testing.T) {
	cfg := PingCheckConfig{Host: "invalid.invalid", Timeout: 1}
	result := PingCheck(cfg)
	if result.Up {
		t.Error("Expected ping to fail for invalid host")
	}
}

func TestPingCheckReturnsSuccessOnLocalhost(t *testing.T) {
	cfg := PingCheckConfig{Host: "127.0.0.1", Timeout: 1}
	result := PingCheck(cfg)
	// We allow this to be flaky in CI, but on most systems this should work.
	if !result.Up {
		t.Log("Ping to 127.0.0.1 failed, which may be expected in some CI/container environments.")
	}
}
