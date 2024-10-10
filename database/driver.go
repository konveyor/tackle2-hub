package database

import (
	"context"
	"database/sql/driver"
	"strings"
	"sync"

	"github.com/mattn/go-sqlite3"
)

// Driver is a wrapper around the SQLite driver.
// The purpose is to prevent database locked errors using
// a mutex around write operations.
type Driver struct {
	mutex   sync.Mutex
	wrapped driver.Driver
	dsn     string
}

// Open a connection.
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

// OpenConnector opens a connection.
func (d *Driver) OpenConnector(dsn string) (dc driver.Connector, err error) {
	d.dsn = dsn
	dc = d
	return
}

// Connect opens a connection.
func (d *Driver) Connect(context.Context) (conn driver.Conn, err error) {
	conn, err = d.Open(d.dsn)
	return
}

// Driver returns the underlying driver.
func (d *Driver) Driver() driver.Driver {
	return d
}

// Conn is a DB connection.
type Conn struct {
	mutex    *sync.Mutex
	wrapped  driver.Conn
	hasMutex bool
	hasTx    bool
}

// Ping the DB.
func (c *Conn) Ping(ctx context.Context) (err error) {
	if p, cast := c.wrapped.(driver.Pinger); cast {
		err = p.Ping(ctx)
	}
	return
}

// ResetSession reset the connection.
// - Reset the Tx.
// - Release the mutex.
func (c *Conn) ResetSession(ctx context.Context) (err error) {
	defer func() {
		c.hasTx = false
		c.release()
	}()
	if p, cast := c.wrapped.(driver.SessionResetter); cast {
		err = p.ResetSession(ctx)
	}
	return
}

// IsValid returns true when the connection is valid.
// When true, the connection may be reused by the sql package.
func (c *Conn) IsValid() (b bool) {
	b = true
	if p, cast := c.wrapped.(driver.Validator); cast {
		b = p.IsValid()
	}
	return
}

// QueryContext execute a query with context.
func (c *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Rows, err error) {
	if c.needsMutex(query) {
		c.acquire()
		defer c.release()
	}
	if p, cast := c.wrapped.(driver.QueryerContext); cast {
		r, err = p.QueryContext(ctx, query, args)
	}
	return
}

// ExecContext executes an SQL/DDL statement with context.
func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Result, err error) {
	if c.needsMutex(query) {
		c.acquire()
		defer c.release()
	}
	if p, cast := c.wrapped.(driver.ExecerContext); cast {
		r, err = p.ExecContext(ctx, query, args)
	}
	return
}

// Begin a transaction.
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
	c.hasTx = true
	return
}

// BeginTx begins a transaction.
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
	c.hasTx = true
	return
}

// Prepare a statement.
func (c *Conn) Prepare(query string) (stmt driver.Stmt, err error) {
	stmt, err = c.wrapped.Prepare(query)
	stmt = &Stmt{
		conn:    c,
		wrapped: stmt,
		query:   query,
	}
	return
}

// PrepareContext prepares a statement with context.
func (c *Conn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	if p, cast := c.wrapped.(driver.ConnPrepareContext); cast {
		stmt, err = p.PrepareContext(ctx, query)
	} else {
		stmt, err = c.Prepare(query)
	}
	stmt = &Stmt{
		conn:    c,
		wrapped: stmt,
		query:   query,
	}
	return
}

// Close the connection.
func (c *Conn) Close() (err error) {
	err = c.wrapped.Close()
	c.hasMutex = false
	c.release()
	return
}

// needsMutex returns true when the query should is a write operation.
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

// acquire the mutex.
// Since Locks are not reentrant, the mutex is acquired
// only if this connection has not already acquired it.
func (c *Conn) acquire() {
	if !c.hasMutex {
		c.mutex.Lock()
		c.hasMutex = true
	}
}

// release the mutex.
// Released only when:
// - This connection has acquired it
// - Not in a transaction.
func (c *Conn) release() {
	if c.hasMutex && !c.hasTx {
		c.mutex.Unlock()
		c.hasMutex = false
	}
}

// endTx report transaction has ended.
func (c *Conn) endTx() {
	c.hasTx = false
}

// Stmt is a SQL/DDL statement.
type Stmt struct {
	wrapped driver.Stmt
	conn    *Conn
	query   string
}

// Close the statement.
func (s *Stmt) Close() (err error) {
	err = s.wrapped.Close()
	return
}

// NumInput returns the number of (query) input parameters.
func (s *Stmt) NumInput() (n int) {
	n = s.wrapped.NumInput()
	return
}

// Exec executes the statement.
func (s *Stmt) Exec(args []driver.Value) (r driver.Result, err error) {
	if s.needsMutex() {
		s.conn.acquire()
		defer s.conn.release()
	}
	r, err = s.wrapped.Exec(args)
	return
}

// ExecContext executes the statement with context.
func (s *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (r driver.Result, err error) {
	if s.needsMutex() {
		s.conn.acquire()
		defer s.conn.release()
	}
	if p, cast := s.wrapped.(driver.StmtExecContext); cast {
		r, err = p.ExecContext(ctx, args)
	} else {
		r, err = s.Exec(s.values(args))
	}
	return
}

// Query executes a query.
func (s *Stmt) Query(args []driver.Value) (r driver.Rows, err error) {
	if s.needsMutex() {
		s.conn.acquire()
		defer s.conn.release()
	}
	r, err = s.wrapped.Query(args)
	return
}

// QueryContext executes a query.
func (s *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (r driver.Rows, err error) {
	if s.needsMutex() {
		s.conn.acquire()
		defer s.conn.release()
	}
	if p, cast := s.wrapped.(driver.StmtQueryContext); cast {
		r, err = p.QueryContext(ctx, args)
	} else {
		r, err = s.Query(s.values(args))
	}
	return
}

// values converts named-values to values.
func (s *Stmt) values(named []driver.NamedValue) (out []driver.Value) {
	for i := range named {
		out = append(out, named[i].Value)
	}
	return
}

// needsMutex returns true when the query should is a write operation.
func (s *Stmt) needsMutex() (matched bool) {
	matched = s.conn.needsMutex(s.query)
	return
}

// Tx is a transaction.
type Tx struct {
	wrapped driver.Tx
	conn    *Conn
}

// Commit the transaction.
// Releases the mutex.
func (t *Tx) Commit() (err error) {
	defer func() {
		t.conn.endTx()
		t.conn.release()
	}()
	err = t.wrapped.Commit()
	return
}

// Rollback the transaction.
// Releases the mutex.
func (t *Tx) Rollback() (err error) {
	defer func() {
		t.conn.endTx()
		t.conn.release()
	}()
	err = t.wrapped.Rollback()
	return
}
