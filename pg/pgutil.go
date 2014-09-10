package pg

func (p DB) RowExists(query string, binds ...interface{}) (bool, error) {
	rows, err := p.Query(query, binds...)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	exists := rows.Next()
	return exists, rows.Err()
}

func (p DB) CreateDatabase(db string) error {
	_, err := p.Exec("create database " + db)
	return err
}

func (p DB) CreateExtension(ext string) error {
	_, err := p.Exec("create extension " + ext)
	return err
}

func (p DB) CreateUser(user, pass string) error {
	_, err := p.Exec("create user " + user + " password '" + pass + "'")
	return err
}

func (p DB) GrantDBOwner(db, user string) error {
	_, err := p.Exec("alter database " + db + " owner to " + user)
	return err
}

func (p DB) DatabaseExists(db string) (bool, error) {
	return p.RowExists(`select * from pg_database
                                where not datistemplate and datname = $1`,
		db)
}

func (p DB) ExtensionExists(ext string) (bool, error) {
	return p.RowExists(`select * from pg_extension where extname = $1`, ext)
}

func (p DB) UserExists(user string) (bool, error) {
	return p.RowExists(`select * from pg_user where usename = $1`, user)
}
