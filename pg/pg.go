package pg

import (
	"database/sql"
	"strconv"

	"github.com/greensnark/go-sequell/ectx"
	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type ConnSpec struct {
	User, Password string
	Database       string
	Host           string
	Port           int
	SSLMode        string
}

func (c ConnSpec) SpecForDB(db string) ConnSpec {
	copy := c
	copy.Database = db
	return copy
}

func (c ConnSpec) DBHost() string {
	if c.Host == "" {
		return "localhost"
	}
	return c.Host
}

func (c ConnSpec) GetSSLMode() string {
	if c.SSLMode == "" {
		return "disable"
	}
	return c.SSLMode
}

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

func (c ConnSpec) Open() (DB, error) {
	cs := c.ConnectionString()
	dbh, err := sql.Open("postgres", cs)
	if err != nil {
		return DB{}, ectx.Err("connect db="+c.Database, err)
	}
	return DB{dbh}, nil
}

func OpenDBUser(db, user, password string) (DB, error) {
	return ConnSpec{Database: db, User: user, Password: password}.Open()
}

type PgError struct {
	Context string
	Cause   error
}

func (p PgError) Error() string {
	return p.Context + ": " + p.Cause.Error()
}

func ctx(context string, err error) error {
	return PgError{Context: context, Cause: err}
}
