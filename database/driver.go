package database

import (
	"context"
	"database/sql/driver"
	"strings"
	"sync"
	"time"

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
		err = withRetry(func() (err error) {
			r, err = p.QueryContext(ctx, query, args)
			return
		})
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
		err = withRetry(func() (err error) {
			r, err = p.ExecContext(ctx, query, args)
			return
		})
	}
	return
}

// Begin a transaction.
func (c *Conn) Begin() (tx driver.Tx, err error) {
	c.acquire()
	err = withRetry(func() (err error) {
		tx, err = c.wrapped.Begin()
		return
	})
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
		err = withRetry(func() (err error) {
			tx, err = p.BeginTx(ctx, opts)
			return
		})
	} else {
		err = withRetry(func() (err error) {
			tx, err = c.wrapped.Begin()
			return
		})
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
	err = withRetry(func() (err error) {
		stmt, err = c.wrapped.Prepare(query)
		return
	})
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
		err = withRetry(func() (err error) {
			stmt, err = p.PrepareContext(ctx, query)
			return
		})
	} else {
		err = withRetry(func() (err error) {
			stmt, err = c.Prepare(query)
			return
		})
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
func (c *Conn) needsMutex(stmt string) (matched bool) {
	stmt = strings.TrimSpace(stmt)
	if stmt != "" {
		action := strings.Fields(stmt)[0]
		action = strings.ToUpper(stmt)
		matched = action != "SELECT"
	}
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
	err = withRetry(func() (err error) {
		r, err = s.wrapped.Exec(args)
		return
	})
	return
}

// ExecContext executes the statement with context.
func (s *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (r driver.Result, err error) {
	if s.needsMutex() {
		s.conn.acquire()
		defer s.conn.release()
	}
	if p, cast := s.wrapped.(driver.StmtExecContext); cast {
		err = withRetry(func() (err error) {
			r, err = p.ExecContext(ctx, args)
			return
		})
	} else {
		err = withRetry(func() (err error) {
			r, err = s.Exec(s.values(args))
			return
		})
	}
	return
}

// Query executes a query.
func (s *Stmt) Query(args []driver.Value) (r driver.Rows, err error) {
	if s.needsMutex() {
		s.conn.acquire()
		defer s.conn.release()
	}
	err = withRetry(func() (err error) {
		r, err = s.wrapped.Query(args)
		return
	})
	return
}

// QueryContext executes a query.
func (s *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (r driver.Rows, err error) {
	if s.needsMutex() {
		s.conn.acquire()
		defer s.conn.release()
	}
	if p, cast := s.wrapped.(driver.StmtQueryContext); cast {
		err = withRetry(func() (err error) {
			r, err = p.QueryContext(ctx, args)
			return
		})
	} else {
		err = withRetry(func() (err error) {
			r, err = s.Query(s.values(args))
			return
		})
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
	err = withRetry(func() (err error) {
		err = t.wrapped.Commit()
		return
	})
	return
}

// Rollback the transaction.
// Releases the mutex.
func (t *Tx) Rollback() (err error) {
	defer func() {
		t.conn.endTx()
		t.conn.release()
	}()
	err = withRetry(func() (err error) {
		err = t.wrapped.Rollback()
		return
	})
	return
}

// withRetry
func withRetry(fn func() error) (err error) {
	retries := 10
	delay := time.Duration(0)
	for i := 0; i < retries; i++ {
		err = fn()
		if err == nil {
			return
		}
		m := err.Error()
		m = strings.ToUpper(m)
		if strings.Contains(m, "LOCKED") || strings.Contains(m, "BUSY") {
			delay += 50 * time.Millisecond
			if delay > time.Second {
				delay = time.Second
			}
			time.Sleep(delay)
		} else {
			break
		}
	}
	return
}
