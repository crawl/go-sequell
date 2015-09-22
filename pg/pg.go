package pg

import (
	"database/sql"
	"strconv"

	"github.com/crawl/go-sequell/ectx"
	_ "github.com/lib/pq" // pg driver for database/sql
)

// DB is a wrapper around database/sql.DB.
type DB struct {
	*sql.DB
}

// ConnSpec is a PostgreSQL connection spec.
type ConnSpec struct {
	User, Password string
	Database       string
	Host           string
	Port           int
	SSLMode        string
}

// SpecForDB clones this connection spec, replacing Database with the given db.
func (c ConnSpec) SpecForDB(db string) ConnSpec {
	copy := c
	copy.Database = db
	return copy
}

// DBHost returns the database host, defaulting to localhost if unspecified.
func (c ConnSpec) DBHost() string {
	if c.Host == "" {
		return "localhost"
	}
	return c.Host
}

// GetSSLMode returns the SSL mode setting, defaulting to "disable".
func (c ConnSpec) GetSSLMode() string {
	if c.SSLMode == "" {
		return "disable"
	}
	return c.SSLMode
}

// ConnectionString constructs a connection string suitable for use in sql.Open
func (c ConnSpec) ConnectionString() string {
	connstr := "sslmode=" + c.GetSSLMode()
	if c.Database != "" {
		connstr += " dbname=" + c.Database
	}
	if c.User != "" {
		connstr += " user=" + c.User
		if c.Password != "" {
			connstr += " password=" + c.Password
		}
	}
	if c.Host != "" {
		connstr += " host=" + c.Host
	}
	if c.Port > 0 {
		connstr += " port=" + strconv.Itoa(c.Port)
	}
	return connstr
}

// Open opens the PostgreSQL database and returns the DB object.
func (c ConnSpec) Open() (DB, error) {
	cs := c.ConnectionString()
	dbh, err := sql.Open("postgres", cs)
	if err != nil {
		return DB{}, ectx.Err("connect db="+c.Database, err)
	}
	return DB{dbh}, nil
}

// OpenDBUser opens the named PostgreSQL database with the given username
// and password.
func OpenDBUser(db, user, password string) (DB, error) {
	return ConnSpec{Database: db, User: user, Password: password}.Open()
}

// Error describes a PostgreSQL error.
type Error struct {
	Context string
	Cause   error
}

// Error returns a string representation of the error object.
func (p Error) Error() string {
	return p.Context + ": " + p.Cause.Error()
}

func ctx(context string, err error) error {
	return Error{Context: context, Cause: err}
}
