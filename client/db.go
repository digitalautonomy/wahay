package client

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	// We’re loading the driver anonymously, aliasing its package
	// qualifier to _ so none of its exported names are visible to
	// our code. Under the hood, the driver registers itself as being
	// available to the database/sql package, but in general nothing
	// else happens with the exception that the init function is run.
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type conn struct {
	filename string
	db       *sql.DB
}

type params map[string]string

func (c *conn) prepareKeysValues(p1 params) (p []string, v []string) {
	for k, v1 := range p1 {
		p = append(p, fmt.Sprintf("`%s`", k))
		v = append(v, fmt.Sprintf("'%s'", v1))
	}

	return
}

func (c *conn) replace(table string, p1 map[string]string) error {
	p, v := c.prepareKeysValues(p1)

	log.Println(fmt.Sprintf(
		"REPLACE INTO `%s` (%s) VALUES (%s)",
		table,
		strings.Join(p, ","),
		strings.Join(v, ","),
	))

	r, err := c.db.Exec(fmt.Sprintf(
		"REPLACE INTO `%s` (%s) VALUES (%s)",
		table,
		strings.Join(p, ","),
		strings.Join(v, ","),
	))
	if err != nil {
		return err
	}

	if n, err := r.RowsAffected(); n == 0 || err != nil {
		return errors.New("sql: replace query did not worked")
	}

	return nil
}

func (c *conn) exists(table string, key, value string) bool {
	var total int

	for r, err := c.db.Query(fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE %s = `%s` LIMIT 1", table, key, value)); err == nil; r.Next() {
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
func getSQLConnection(sqlFile string) (*conn, error) {
	c := &conn{}

	if !fileExists(sqlFile) {
		return c, errors.New("the database doesn't exists")
	}

	var err error
	c.filename = sqlFile

	c.db, err = sql.Open("sqlite3", sqlFile)
	if err != nil {
		return nil, err
	}

	return c, nil
}
