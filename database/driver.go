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
	c.release()
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
	if c.needsMutex(query) {
		c.acquire()
	}
	if p, cast := c.wrapped.(driver.QueryerContext); cast {
		r, err = p.QueryContext(ctx, query, args)
	}
	return
}

func (c *Conn) PrepareContext(ctx context.Context, query string) (s driver.Stmt, err error) {
	if c.needsMutex(query) {
		c.acquire()
	}
	if p, cast := c.wrapped.(driver.ConnPrepareContext); cast {
		s, err = p.PrepareContext(ctx, query)
	}
	return
}

func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Result, err error) {
	c.acquire()
	if p, cast := c.wrapped.(driver.ExecerContext); cast {
		r, err = p.ExecContext(ctx, query, args)
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
	return
}

func (c *Conn) Prepare(query string) (s driver.Stmt, err error) {
	if c.needsMutex(query) {
		c.acquire()
	}
	s, err = c.wrapped.Prepare(query)
	return
}

func (c *Conn) Close() (err error) {
	err = c.wrapped.Close()
	c.release()
	return
}

func (c *Conn) Begin() (tx driver.Tx, err error) {
	c.acquire()
	tx, err = c.wrapped.Begin()
	if err != nil {
		return
	}
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
