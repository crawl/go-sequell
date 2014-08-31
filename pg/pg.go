package pg

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type PgDB struct {
	*sql.DB
}

func OpenDB(db, user, password string) (PgDB, error) {
	dbh, err :=
		sql.Open(
			"postgres",
			"sslmode=disable user="+user+" password="+password+" dbname="+db)
	if err != nil {
		return PgDB{}, err
	}
	return PgDB{dbh}, nil
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
