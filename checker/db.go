package checker

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
	db, err := sql.Open(string(cfg.Driver), cfg.DSN)
	if err != nil {
		return DBResult{Up: false, RespTime: 0, Err: err}
	}
	defer db.Close()
	ch := make(chan error, 1)
	go func() {
		ch <- db.Ping()
	}()
	select {
	case err := <-ch:
		respTime := time.Since(start).Seconds()
		return DBResult{Up: err == nil, RespTime: respTime, Err: err}
	case <-time.After(cfg.Timeout):
		return DBResult{Up: false, RespTime: time.Since(start).Seconds(), Err: sql.ErrConnDone}
	}
}
