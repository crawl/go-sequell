package pg

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type ConnSpec struct {
	User, Password string
	Database       string
}

func (c ConnSpec) Open() (DB, error) {
	dbh, err :=
		sql.Open(
			"postgres",
			"sslmode=disable user="+c.User+
				" password="+c.Password+
				" dbname="+c.Database)
	if err != nil {
		return DB{}, err
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
