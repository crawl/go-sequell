package pg

// RowExists checks if query returns any rows
func (p DB) RowExists(query string, binds ...interface{}) (bool, error) {
	rows, err := p.Query(query, binds...)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	exists := rows.Next()
	return exists, rows.Err()
}

// CreateDatabase creates a database named db
func (p DB) CreateDatabase(db string) error {
	_, err := p.Exec("create database " + db)
	return err
}

// CreateExtension creates a pg extension named ext
func (p DB) CreateExtension(ext string) error {
	_, err := p.Exec("create extension " + ext)
	return err
}

// CreateUser creates a user with the given password pass.
func (p DB) CreateUser(user, pass string) error {
	_, err := p.Exec("create user " + user + " password '" + pass + "'")
	return err
}

// GrantDBOwner alters the database owner to user.
func (p DB) GrantDBOwner(db, user string) error {
	_, err := p.Exec("alter database " + db + " owner to " + user)
	return err
}

// DatabaseExists checks if the database named db exists.
func (p DB) DatabaseExists(db string) (bool, error) {
	return p.RowExists(`select * from pg_database
                                where not datistemplate and datname = $1`,
		db)
}

// ExtensionExists checks if the postgres extension ext exists.
func (p DB) ExtensionExists(ext string) (bool, error) {
	return p.RowExists(`select * from pg_extension where extname = $1`, ext)
}

// UserExists checks if the user exists.
func (p DB) UserExists(user string) (bool, error) {
	return p.RowExists(`select * from pg_user where usename = $1`, user)
}
