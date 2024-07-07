package database

import (
	"context"
	"database/sql/driver"
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
	if p, cast := c.wrapped.(driver.QueryerContext); cast {
		r, err = p.QueryContext(ctx, query, args)
	}
	return
}

func (c *Conn) PrepareContext(ctx context.Context, query string) (s driver.Stmt, err error) {
	if p, cast := c.wrapped.(driver.ConnPrepareContext); cast {
		s, err = p.PrepareContext(ctx, query)
	}
	return
}

func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Result, err error) {
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
	c.mutex.Lock()
	return
}

func (c *Conn) Prepare(query string) (s driver.Stmt, err error) {
	s, err = c.wrapped.Prepare(query)
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
