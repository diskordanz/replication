package rep

import (
	"database/sql"
	"strings"
	"sync/atomic"
)

type DB struct {
	pdbs  []*sql.DB
	count uint64
}

func Open(driverName, dataSourceNames string) (*DB, error) {
	conns := strings.Split(dataSourceNames, ";")
	db := &DB{pdbs: make([]*sql.DB, len(conns))}
	err := scatter(len(db.pdbs), func(i int) (err error) {
		db.pdbs[i], err = sql.Open(driverName, conns[i])
		return err
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error {
	return scatter(len(db.pdbs), func(i int) error {
		return db.pdbs[i].Close()
	})
}

func (db *DB) Exec(query string, args ...interface{}) error {
	scatter(len(db.pdbs), func(i int) error {
		db.pdbs[i].Exec(query, args...)
		return nil
	})
	return nil
}

func (db *DB) Sync(query string, args ...interface{}) error {
	scatter(len(db.pdbs)-1, func(i int) error {
		db.pdbs[i+1].Exec(query, args...)
		return nil
	})
	return nil
}

func (db *DB) Ping() error {
	return scatter(len(db.pdbs), func(i int) error {
		return db.pdbs[i].Ping()
	})
}
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.pdbs[db.slave(len(db.pdbs))].Query(query, args...)
}
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.pdbs[db.slave(len(db.pdbs))].QueryRow(query, args...)
}

func (db *DB) Slave() *sql.DB {
	return db.pdbs[db.slave(len(db.pdbs))]
}
func (db *DB) Master() *sql.DB {
	return db.pdbs[0]
}

func (db *DB) slave(n int) int {
	if n <= 1 {
		return 0
	}
	return int(1 + (atomic.AddUint64(&db.count, 1) % uint64(n-1)))
}

func scatter(n int, fn func(i int) error) error {
	errors := make(chan error, n)

	var i int
	for i = 0; i < n; i++ {
		go func(i int) { errors <- fn(i) }(i)
	}

	var err, innerErr error
	for i = 0; i < cap(errors); i++ {
		if innerErr = <-errors; innerErr != nil {
			err = innerErr
		}
	}

	return err
}

type Stmt interface {
	Close() error
	Exec(...interface{}) (sql.Result, error)
	Query(...interface{}) (*sql.Rows, error)
	QueryRow(...interface{}) *sql.Row
}
type stmt struct {
	db    *DB
	stmts []*sql.Stmt
}

func (s *stmt) Close() error {
	return scatter(len(s.stmts), func(i int) error {
		return s.stmts[i].Close()
	})
}
func (s *stmt) Exec(args ...interface{}) (sql.Result, error) {
	return s.stmts[0].Exec(args...)
}

func (s *stmt) Query(args ...interface{}) (*sql.Rows, error) {
	return s.stmts[s.db.slave(len(s.db.pdbs))].Query(args...)
}

func (s *stmt) QueryRow(args ...interface{}) *sql.Row {
	return s.stmts[s.db.slave(len(s.db.pdbs))].QueryRow(args...)
}
