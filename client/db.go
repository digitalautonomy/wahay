package client

import (
	"database/sql"
	"errors"

	// The sqlite3 driver is automatic added by this package
	_ "github.com/mattn/go-sqlite3"
)

type conn struct {
	filename string
	db       *sql.DB
}

func (c *conn) replace(query string, params ...interface{}) error {
	r, err := c.db.Exec(query, params...)
	if err != nil {
		return err
	}

	if n, err := r.RowsAffected(); n == 0 || err != nil {
		return errors.New("sql: replace query did not worked")
	}

	return nil
}

const queryExistsText = "SELECT COUNT(*) FROM `?` WHERE `?` = ? LIMIT 1"

func (c *conn) exists(table string, key, value string) bool {
	var total int
	for r, err := c.db.Query(queryExistsText, table, key, value); err == nil; r.Next() {
		err = r.Scan(&total)
		if err != nil {
			return false
		}
	}

	return total == 1
}

func (c *conn) close() error {
	return c.db.Close()
}

// The caller should close the connection
func getSQLConnection(filename string) (*conn, error) {
	c := &conn{}

	if !fileExists(filename) {
		return c, errors.New("the database doesn't exists")
	}

	var err error
	c.filename = filename

	c.db, err = sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	return c, nil
}
