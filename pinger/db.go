package pinger

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type DBType string

const (
	Postgres DBType = "postgres"
	MySQL    DBType = "mysql"
	MariaDB  DBType = "mariadb"
	Redis    DBType = "redis"
)

type DBCheckConfig struct {
	Name     string
	DSN      string // Data Source Name
	Driver   DBType
	Timeout  time.Duration
}

type DBResult struct {
	Up       bool
	RespTime float64
	Err      error
}

func DBCheck(cfg DBCheckConfig) DBResult {
	if cfg.Driver == Redis {
		return RedisCheck(cfg)
	}
	start := time.Now()
	driver := string(cfg.Driver)
	if cfg.Driver == MariaDB {
		// MariaDB is wire-compatible with MySQL
		driver = string(MySQL)
	}
	log.Printf("[DBCheck] Opening DB connection for %s using driver %s", cfg.Name, driver)
	db, err := sql.Open(driver, cfg.DSN)
	if err != nil {
		log.Printf("[DBCheck] Failed to open DB connection: %v", err)
		return DBResult{Up: false, RespTime: 0, Err: err}
	}
	defer func() {
		log.Printf("[DBCheck] Closing DB connection for %s", cfg.Name)
		db.Close()
	}()
	ch := make(chan error, 1)
	go func() {
		ch <- db.Ping()
	}()
	select {
	case err := <-ch:
		respTime := time.Since(start).Seconds()
		if err != nil {
			log.Printf("[DBCheck] Ping failed for %s: %v", cfg.Name, err)
			if err.Error() != "" && (containsSSL(err.Error()) || containsTLS(err.Error())) {
				log.Printf("[DBCheck] SSL/TLS error details for %s: %v", cfg.Name, err)
			}
		} else {
			log.Printf("[DBCheck] Ping succeeded for %s", cfg.Name)
		}
		return DBResult{Up: err == nil, RespTime: respTime, Err: err}
	case <-time.After(cfg.Timeout):
		respTime := time.Since(start).Seconds()
		log.Printf("[DBCheck] Timeout after %.2fs for %s", respTime, cfg.Name)
		return DBResult{Up: false, RespTime: respTime, Err: sql.ErrConnDone}
	}
}

// RedisCheck pings a Redis instance. The DSN may be a redis:// or rediss://
// (TLS) URL, a unix:// socket URL, or a bare host:port.
func RedisCheck(cfg DBCheckConfig) DBResult {
	start := time.Now()
	dsn := cfg.DSN
	if !strings.Contains(dsn, "://") {
		dsn = "redis://" + dsn
	}
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		log.Printf("[DBCheck] Failed to parse Redis DSN for %s: %v", cfg.Name, err)
		return DBResult{Up: false, RespTime: 0, Err: err}
	}
	// Retries are handled by the check runner; -1 disables client retries
	// so failure detection stays fast and the logs quiet.
	opts.MaxRetries = -1
	log.Printf("[DBCheck] Opening Redis connection for %s to %s", cfg.Name, opts.Addr)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	client := redis.NewClient(opts)
	defer func() {
		log.Printf("[DBCheck] Closing Redis connection for %s", cfg.Name)
		client.Close()
	}()
	err = client.Ping(ctx).Err()
	respTime := time.Since(start).Seconds()
	if err != nil {
		log.Printf("[DBCheck] Redis ping failed for %s: %v", cfg.Name, err)
		return DBResult{Up: false, RespTime: respTime, Err: err}
	}
	log.Printf("[DBCheck] Redis ping succeeded for %s", cfg.Name)
	return DBResult{Up: true, RespTime: respTime, Err: nil}
}

// containsSSL checks if the error string contains 'ssl:'
func containsSSL(s string) bool {
	return len(s) > 4 && s[:4] == "ssl:"
}
