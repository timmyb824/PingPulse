package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	upGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "uptime_check_up",
			Help: "Whether the target is up (1) or down (0)",
		},
		[]string{"type", "name"},
	)
	respTimeGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "uptime_check_response_seconds",
			Help: "Response time in seconds",
		},
		[]string{"type", "name"},
	)
	sslExpiryGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "uptime_check_ssl_days_left",
			Help: "Days left until SSL cert expiry",
		},
		[]string{"name"},
	)
	successCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uptime_check_success_total",
			Help: "Total successful checks",
		},
		[]string{"type", "name"},
	)
	failureCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uptime_check_failure_total",
			Help: "Total failed checks",
		},
		[]string{"type", "name"},
	)
	sslErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "uptime_check_ssl_errors_total",
			Help: "Total number of SSL errors encountered during HTTP checks.",
		},
		[]string{"name", "error_type"},
	)
)
