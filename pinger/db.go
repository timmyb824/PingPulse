package pinger

import (
	"database/sql"
	_ "github.com/lib/pq"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type DBType string

const (
	Postgres DBType = "postgres"
	MySQL    DBType = "mysql"
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
	start := time.Now()
	log.Printf("[DBCheck] Opening DB connection for %s using driver %s", cfg.Name, cfg.Driver)
	db, err := sql.Open(string(cfg.Driver), cfg.DSN)
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

// containsSSL checks if the error string contains 'ssl:'
func containsSSL(s string) bool {
	return len(s) > 4 && s[:4] == "ssl:"
}
