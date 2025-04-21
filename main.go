package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/timmyb824/httping/config"
	"github.com/timmyb824/httping/checker"
)

// For simplicity, metrics are defined here in main.go. For larger codebases, consider moving these to a separate metrics.go.
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
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: httping <config.yaml>")
		os.Exit(1)
	}
	configPath := os.Args[1]
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	prometheus.MustRegister(upGauge, respTimeGauge, sslExpiryGauge, successCounter, failureCounter)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Prometheus metrics at :8080/metrics")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// --- Graceful config reload ---
	reload := make(chan struct{}, 1)
	go watchConfigFile(configPath, reload)

	interval := time.Duration(cfg.IntervalSeconds) * time.Second
	if interval == 0 {
		interval = 30 * time.Second
	}

	for {
		select {
		case <-reload:
			log.Println("Reloading config...")
			newCfg, err := config.LoadConfig(configPath)
			if err != nil {
				log.Printf("Failed to reload config: %v (keeping old config)", err)
			} else {
				cfg = newCfg
				interval = time.Duration(cfg.IntervalSeconds) * time.Second
				if interval == 0 {
					interval = 30 * time.Second
				}
			}
		default:
		}
		if cfg.MaintenanceMode {
			setMaintenanceMetrics(cfg)
			log.Println("MAINTENANCE MODE: All checks skipped.")
			time.Sleep(interval)
			continue
		}
		var wg sync.WaitGroup
		for _, hc := range cfg.HTTPChecks {
			wg.Add(1)
			go func(hc config.HTTPCheck) {
				defer wg.Done()
				log.Printf("[HTTP] Starting check: %s (%s)", hc.Name, hc.URL)
				result := checker.HTTPCheck(checker.HTTPCheckConfig{
					URL: hc.URL,
					Timeout: time.Duration(hc.Timeout) * time.Second,
					AcceptStatusCodes: hc.AcceptStatusCodes,
				})
				up := 0.0
				if result.Up {
					up = 1.0
					successCounter.WithLabelValues("http", hc.Name).Inc()
					log.Printf("[HTTP] SUCCESS: %s | status=%d, resp=%.3fs, ssl_days=%d", hc.Name, result.StatusCode, result.RespTime, result.SSLDaysLeft)
				} else {
					failureCounter.WithLabelValues("http", hc.Name).Inc()
					log.Printf("[HTTP] FAIL: %s | status=%d, resp=%.3fs, ssl_days=%d, err=%v", hc.Name, result.StatusCode, result.RespTime, result.SSLDaysLeft, result.Err)
				}
				upGauge.WithLabelValues("http", hc.Name).Set(up)
				respTimeGauge.WithLabelValues("http", hc.Name).Set(result.RespTime)
				if result.SSLDaysLeft >= 0 {
					sslExpiryGauge.WithLabelValues(hc.Name).Set(float64(result.SSLDaysLeft))
				}
			}(hc)
		}
		for _, pc := range cfg.PingChecks {
			wg.Add(1)
			go func(pc config.PingCheck) {
				defer wg.Done()
				log.Printf("[PING] Starting check: %s (%s)", pc.Name, pc.Host)
				result := checker.PingCheck(checker.PingCheckConfig{
					Host: pc.Host,
					Timeout: time.Duration(pc.Timeout) * time.Second,
				})
				up := 0.0
				if result.Up {
					up = 1.0
					successCounter.WithLabelValues("ping", pc.Name).Inc()
					log.Printf("[PING] SUCCESS: %s | resp=%.3fs", pc.Name, result.RespTime)
				} else {
					failureCounter.WithLabelValues("ping", pc.Name).Inc()
					log.Printf("[PING] FAIL: %s | resp=%.3fs, err=%v", pc.Name, result.RespTime, result.Err)
				}
				upGauge.WithLabelValues("ping", pc.Name).Set(up)
				respTimeGauge.WithLabelValues("ping", pc.Name).Set(result.RespTime)
			}(pc)
		}
		for _, dbc := range cfg.DBChecks {
			wg.Add(1)
			go func(dbc config.DBCheck) {
				defer wg.Done()
				log.Printf("[DB] Starting check: %s (driver=%s)", dbc.Name, dbc.Driver)
				result := checker.DBCheck(checker.DBCheckConfig{
					Name: dbc.Name,
					Driver: checker.DBType(dbc.Driver),
					DSN: dbc.DSN,
					Timeout: time.Duration(dbc.Timeout) * time.Second,
				})
				up := 0.0
				if result.Up {
					up = 1.0
					successCounter.WithLabelValues("db", dbc.Name).Inc()
					log.Printf("[DB] SUCCESS: %s | resp=%.3fs", dbc.Name, result.RespTime)
				} else {
					failureCounter.WithLabelValues("db", dbc.Name).Inc()
					log.Printf("[DB] FAIL: %s | resp=%.3fs, err=%v", dbc.Name, result.RespTime, result.Err)
				}
				upGauge.WithLabelValues("db", dbc.Name).Set(up)
				respTimeGauge.WithLabelValues("db", dbc.Name).Set(result.RespTime)
			}(dbc)
		}
		wg.Wait()
		log.Println("Checks complete. Sleeping...", interval)
		time.Sleep(interval)
	}
}

func setMaintenanceMetrics(cfg *config.Config) {
	for _, hc := range cfg.HTTPChecks {
		upGauge.WithLabelValues("http", hc.Name).Set(0)
		respTimeGauge.WithLabelValues("http", hc.Name).Set(0)
		sslExpiryGauge.WithLabelValues(hc.Name).Set(0)
	}
	for _, pc := range cfg.PingChecks {
		upGauge.WithLabelValues("ping", pc.Name).Set(0)
		respTimeGauge.WithLabelValues("ping", pc.Name).Set(0)
	}
	for _, dbc := range cfg.DBChecks {
		upGauge.WithLabelValues("db", dbc.Name).Set(0)
		respTimeGauge.WithLabelValues("db", dbc.Name).Set(0)
	}
}

// watchConfigFile watches the config file for changes and signals reloads.
func watchConfigFile(path string, reload chan<- struct{}) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("fsnotify error: %v", err)
		return
	}
	defer watcher.Close()
	_ = watcher.Add(path)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				log.Printf("Config file changed: %v", event)
				reload <- struct{}{}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("fsnotify error: %v", err)
		}
	}
}
