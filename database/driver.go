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
	mutex    *sync.Mutex
	wrapped  driver.Conn
	hasMutex bool
}

func (c *Conn) Ping(ctx context.Context) (err error) {
	if p, cast := c.wrapped.(driver.Pinger); cast {
		err = p.Ping(ctx)
	}
	return
}

func (c *Conn) ResetSession(ctx context.Context) (err error) {
	defer c.release()
	if p, cast := c.wrapped.(driver.SessionResetter); cast {
		err = p.ResetSession(ctx)
	}
	return
}
func (c *Conn) IsValid() (b bool) {
	b = true
	if p, cast := c.wrapped.(driver.Validator); cast {
		b = p.IsValid()
	}
	return
}

func (c *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Rows, err error) {
	defer c.release()
	if c.needsMutex(query) {
		c.acquire()
	}
	if p, cast := c.wrapped.(driver.QueryerContext); cast {
		r, err = p.QueryContext(ctx, query, args)
	}
	return
}

func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Result, err error) {
	defer c.release()
	if c.needsMutex(query) {
		c.acquire()
	}
	if p, cast := c.wrapped.(driver.ExecerContext); cast {
		r, err = p.ExecContext(ctx, query, args)
	}
	return
}

func (c *Conn) Begin() (tx driver.Tx, err error) {
	c.acquire()
	tx, err = c.wrapped.Begin()
	if err != nil {
		return
	}
	tx = &Tx{
		conn:    c,
		wrapped: tx,
	}
	return
}

func (c *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	c.acquire()
	if p, cast := c.wrapped.(driver.ConnBeginTx); cast {
		tx, err = p.BeginTx(ctx, opts)
	} else {
		tx, err = c.wrapped.Begin()
	}
	tx = &Tx{
		conn:    c,
		wrapped: tx,
	}
	return
}

func (c *Conn) Prepare(query string) (stmt driver.Stmt, err error) {
	if c.needsMutex(query) {
		c.acquire()
	}
	stmt, err = c.wrapped.Prepare(query)
	stmt = &Stmt{
		conn:    c,
		wrapped: stmt,
	}
	return
}

func (c *Conn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	if c.needsMutex(query) {
		c.acquire()
	}
	if p, cast := c.wrapped.(driver.ConnPrepareContext); cast {
		stmt, err = p.PrepareContext(ctx, query)
	} else {
		stmt, err = c.Prepare(query)
	}
	stmt = &Stmt{
		conn:    c,
		wrapped: stmt,
	}
	return
}

func (c *Conn) Close() (err error) {
	err = c.wrapped.Close()
	c.release()
	return
}

func (c *Conn) needsMutex(query string) (matched bool) {
	if query == "" {
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

func (c *Conn) acquire() {
	if !c.hasMutex {
		c.mutex.Lock()
		c.hasMutex = true
	}
}

func (c *Conn) release() {
	if c.hasMutex {
		c.mutex.Unlock()
		c.hasMutex = false
	}
}

type Stmt struct {
	wrapped driver.Stmt
	conn    *Conn
}

func (s *Stmt) Close() (err error) {
	defer s.conn.release()
	err = s.wrapped.Close()
	return
}

func (s *Stmt) NumInput() (n int) {
	n = s.wrapped.NumInput()
	return
}
func (s *Stmt) Exec(args []driver.Value) (r driver.Result, err error) {
	defer s.conn.release()
	r, err = s.wrapped.Exec(args)
	return
}

func (s *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (r driver.Result, err error) {
	defer s.conn.release()
	if p, cast := s.wrapped.(driver.StmtExecContext); cast {
		r, err = p.ExecContext(ctx, args)
	} else {
		r, err = s.Exec(s.values(args))
	}
	return
}

func (s *Stmt) Query(args []driver.Value) (r driver.Rows, err error) {
	r, err = s.wrapped.Query(args)
	r = &Rows{
		conn:    s.conn,
		wrapped: r,
	}
	return
}

func (s *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (r driver.Rows, err error) {
	if p, cast := s.wrapped.(driver.StmtQueryContext); cast {
		r, err = p.QueryContext(ctx, args)
	} else {
		r, err = s.Query(s.values(args))
	}
	r = &Rows{
		conn:    s.conn,
		wrapped: r,
	}
	return
}

func (s *Stmt) values(named []driver.NamedValue) (out []driver.Value) {
	for i := range named {
		out = append(out, named[i].Value)
	}
	return
}

type Tx struct {
	wrapped driver.Tx
	conn    *Conn
}

func (t *Tx) Commit() (err error) {
	defer t.conn.release()
	err = t.wrapped.Commit()
	return
}

func (t *Tx) Rollback() (err error) {
	defer t.conn.release()
	err = t.wrapped.Rollback()
	return
}

type Rows struct {
	conn    *Conn
	wrapped driver.Rows
}

func (r *Rows) Columns() (s []string) {
	s = r.wrapped.Columns()
	return
}

func (r *Rows) Close() (err error) {
	defer r.conn.release()
	err = r.wrapped.Close()
	return
}

func (r *Rows) Next(object []driver.Value) (err error) {
	err = r.wrapped.Next(object)
	return
}
