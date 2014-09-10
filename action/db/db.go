package db

import (
	"fmt"
	"os"

	"github.com/greensnark/go-sequell/ectx"
	"github.com/greensnark/go-sequell/pg"
	"github.com/greensnark/go-sequell/schema"
)

var DbExtensions = []string{"citext", "orafce"}

func DumpSchema(dbspec pg.ConnSpec) error {
	db, err := dbspec.Open()
	if err != nil {
		return err
	}
	s, err := db.IntrospectSchema()
	if err != nil {
		return err
	}
	s.Sort().Write(schema.SelTablesIndexes, os.Stdout)
	return nil
}

func CreateDB(admin, db pg.ConnSpec) error {
	pgdb, err := admin.Open()
	if err != nil {
		return err
	}
	defer pgdb.Close()
	dbexist, err := pgdb.DatabaseExists(db.Database)
	if err != nil {
		return err
	}
	if !dbexist {
		fmt.Println("Creating database", db.Database)
		if err = pgdb.CreateDatabase(db.Database); err != nil {
			return ectx.Err("CreateDatabase", err)
		}
	}

	if err = CreateExtensions(admin.SpecForDB(db.Database)); err != nil {
		return ectx.Err("CreateExtensions", err)
	}

	if err = CreateUser(pgdb, db); err != nil {
		return ectx.Err("CreateUser", err)
	}
	return ectx.Err("GrantDBOwner", pgdb.GrantDBOwner(db.Database, db.User))
}

func CreateUser(pgdb pg.DB, dbspec pg.ConnSpec) error {
	userExist, err := pgdb.UserExists(dbspec.User)
	if err != nil {
		return err
	}
	if !userExist {
		if err = pgdb.CreateUser(dbspec.User, dbspec.Password); err != nil {
			return err
		}
	}
	return nil
}

func CreateExtensions(db pg.ConnSpec) error {
	c, err := db.Open()
	if err != nil {
		return err
	}
	defer c.Close()
	for _, ext := range DbExtensions {
		extExists, err := c.ExtensionExists(ext)
		if err != nil {
			return err
		}
		if !extExists {
			if err = c.CreateExtension(ext); err != nil {
				return err
			}
		}
	}
	return nil
}
