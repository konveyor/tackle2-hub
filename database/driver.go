package database

import (
	"context"
	"database/sql/driver"
	"strings"
	"sync"

	"github.com/mattn/go-sqlite3"
)

type Driver struct {
	mutex   sync.Mutex
	wrapped driver.Driver
	dsn     string
}

func (d *Driver) Open(dsn string) (conn driver.Conn, err error) {
	d.wrapped = &sqlite3.SQLiteDriver{}
	conn, err = d.wrapped.Open(dsn)
	if err != nil {
		return
	}
	conn = &Conn{
		mutex:   &d.mutex,
		wrapped: conn,
	}
	return
}

func (d *Driver) OpenConnector(dsn string) (dc driver.Connector, err error) {
	d.dsn = dsn
	dc = d
	return
}

func (d *Driver) Connect(context.Context) (conn driver.Conn, err error) {
	conn, err = d.Open(d.dsn)
	return
}

func (d *Driver) Driver() driver.Driver {
	return d
}

type Conn struct {
	mutex   *sync.Mutex
	wrapped driver.Conn
	tx      driver.Tx
}

func (c *Conn) Ping(ctx context.Context) (err error) {
	if p, cast := c.wrapped.(driver.Pinger); cast {
		err = p.Ping(ctx)
	}
	return
}

func (c *Conn) ResetSession(ctx context.Context) (err error) {
	if p, cast := c.wrapped.(driver.SessionResetter); cast {
		err = p.ResetSession(ctx)
	}
	return
}
func (c *Conn) IsValid() (b bool) {
	if p, cast := c.wrapped.(driver.Validator); cast {
		b = p.IsValid()
	}
	return
}

func (c *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Rows, err error) {
	if c.tx == nil {
		c.mutex.Lock()
		defer c.mutex.Unlock()
	}
	if p, cast := c.wrapped.(driver.QueryerContext); cast {
		r, err = p.QueryContext(ctx, query, args)
	}
	return
}

func (c *Conn) PrepareContext(ctx context.Context, query string) (s driver.Stmt, err error) {
	if p, cast := c.wrapped.(driver.ConnPrepareContext); cast {
		s, err = p.PrepareContext(ctx, query)
	}
	if err != nil {
		return
	}
	stmtLocked := c.stmtLocked(query)
	s = &Stmt{
		mutex:   c.mutex,
		locked:  stmtLocked,
		wrapped: s,
	}
	if stmtLocked {
		c.mutex.Lock()
	}
	return
}

func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Result, err error) {
	if c.tx == nil {
		c.mutex.Lock()
		defer c.mutex.Unlock()
	}
	if p, cast := c.wrapped.(driver.ExecerContext); cast {
		r, err = p.ExecContext(ctx, query, args)
	}
	return
}

func (c *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	if p, cast := c.wrapped.(driver.ConnBeginTx); cast {
		tx, err = p.BeginTx(ctx, opts)
	} else {
		tx, err = c.wrapped.Begin()
	}
	if err != nil {
		return
	}
	tx = &Tx{
		mutex:   c.mutex,
		wrapped: tx,
	}
	c.tx = tx
	c.mutex.Lock()
	return
}

func (c *Conn) Prepare(query string) (s driver.Stmt, err error) {
	s, err = c.wrapped.Prepare(query)
	if err != nil {
		return
	}
	stmtLocked := c.stmtLocked(query)
	s = &Stmt{
		mutex:   c.mutex,
		locked:  stmtLocked,
		wrapped: s,
	}
	if stmtLocked {
		c.mutex.Lock()
	}
	return
}

func (c *Conn) Close() (err error) {
	err = c.wrapped.Close()
	return
}

func (c *Conn) Begin() (tx driver.Tx, err error) {
	tx, err = c.wrapped.Begin()
	if err != nil {
		return
	}
	tx = &Tx{
		mutex:   c.mutex,
		wrapped: tx,
	}
	c.tx = tx
	c.mutex.Lock()
	return
}

func (c *Conn) stmtLocked(query string) (matched bool) {
	if c.tx != nil || query == "" {
		return
	}
	query = strings.ToUpper(query)
	action := strings.Fields(query)[0]
	action = strings.ToUpper(action)
	matched = action == "CREATE" ||
		action == "INSERT" ||
		action == "UPDATE" ||
		action == "DELETE"
	return
}

type Tx struct {
	mutex   *sync.Mutex
	wrapped driver.Tx
}

func (tx *Tx) Commit() (err error) {
	defer func() {
		tx.mutex.Unlock()
	}()
	err = tx.wrapped.Commit()
	return
}
func (tx *Tx) Rollback() (err error) {
	defer func() {
		tx.mutex.Unlock()
	}()
	err = tx.wrapped.Rollback()
	return
}

type Stmt struct {
	mutex   *sync.Mutex
	wrapped driver.Stmt
	locked  bool
}

func (s *Stmt) Close() (err error) {
	if s.locked {
		s.mutex.Unlock()
	}
	err = s.wrapped.Close()
	return
}
func (s *Stmt) NumInput() (n int) {
	n = s.wrapped.NumInput()
	return
}
func (s *Stmt) Exec(args []driver.Value) (r driver.Result, err error) {
	r, err = s.wrapped.Exec(args)
	return
}

func (s *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (r driver.Result, err error) {
	if p, cast := s.wrapped.(driver.StmtExecContext); cast {
		r, err = p.ExecContext(ctx, args)
	}
	return
}

func (s *Stmt) Query(args []driver.Value) (r driver.Rows, err error) {
	r, err = s.wrapped.Query(args)
	return
}

func (s *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (r driver.Rows, err error) {
	if p, cast := s.wrapped.(driver.StmtQueryContext); cast {
		r, err = p.QueryContext(ctx, args)
	}
	return
}
