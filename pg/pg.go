package pg

import (
	"database/sql"

	"github.com/greensnark/go-sequell/ectx"
	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type ConnSpec struct {
	User, Password string
	Database       string
}

func (c ConnSpec) SpecForDB(db string) ConnSpec {
	copy := c
	copy.Database = db
	return copy
}

func (c ConnSpec) ConnectionString() string {
	res := "sslmode=disable"
	if c.Database != "" {
		res += " dbname=" + c.Database
	}
	if c.User != "" {
		res += " user=" + c.User
	}
	if c.Password != "" {
		res += " password=" + c.Password
	}
	return res
}

func (c ConnSpec) Open() (DB, error) {
	dbh, err := sql.Open("postgres", c.ConnectionString())
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
