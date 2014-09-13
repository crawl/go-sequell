package db

import (
	"fmt"
	"os"

	"github.com/greensnark/go-sequell/action"
	"github.com/greensnark/go-sequell/crawl/data"
	cdb "github.com/greensnark/go-sequell/crawl/db"
	"github.com/greensnark/go-sequell/ectx"
	"github.com/greensnark/go-sequell/loader"
	"github.com/greensnark/go-sequell/pg"
	"github.com/greensnark/go-sequell/schema"
	"github.com/greensnark/go-sequell/sources"
)

var DbExtensions = []string{"citext", "orafce"}

func CrawlSchema() *cdb.CrawlSchema {
	schema, err := cdb.LoadSchema(data.Crawl)
	if err != nil {
		panic(err)
	}
	return schema
}

func Sources() *sources.Servers {
	src, err := sources.Sources(data.Sources(), action.LogCache)
	if err != nil {
		panic(err)
	}
	return src
}

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
		fmt.Printf("Creating database \"%s\"\n", db.Database)
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
		fmt.Printf("Creating user \"%s\"\n", dbspec.User)
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
			fmt.Printf("Creating extension \"%s\"\n", ext)
			if err = c.CreateExtension(ext); err != nil {
				return err
			}
		}
	}
	return nil
}

func PrintSchema(skipIndexes, dropIndexes, createIndexes bool) {
	s := CrawlSchema().Schema()
	sel := schema.SelTablesIndexes
	if skipIndexes {
		sel = schema.SelTables
	}
	if dropIndexes {
		sel = schema.SelDropIndexes
	}
	if createIndexes {
		sel = schema.SelIndexes
	}
	s.Sort().Write(sel, os.Stdout)
}

func CheckDBSchema(dbspec pg.ConnSpec, applyDelta bool) error {
	db, err := dbspec.Open()
	if err != nil {
		return err
	}
	actualSchema, err := db.IntrospectSchema()
	if err != nil {
		return err
	}
	wantedSchema := CrawlSchema().Schema()
	diff := wantedSchema.DiffSchema(actualSchema)
	if len(diff.Tables) == 0 {
		fmt.Fprintf(os.Stderr, "Schema is up-to-date.\n")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Schema delta:\n")
	diff.PrintDelta(os.Stderr)
	if applyDelta {
		return nil
	}
	return nil
}

func CreateDBSchema(db pg.ConnSpec) error {
	c, err := db.Open()
	if err != nil {
		return err
	}
	defer c.Close()
	s := CrawlSchema().Schema()
	fmt.Printf("Creating tables in database \"%s\"\n", db.Database)
	for _, sql := range s.SqlSel(schema.SelTables) {
		if _, err = c.Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func DropDB(admin pg.ConnSpec, db pg.ConnSpec, force bool) error {
	if !force {
		return fmt.Errorf("Use --force to drop the database '%s'", db.Database)
	}
	adminDB, err := admin.Open()
	if err != nil {
		return err
	}

	fmt.Printf("Dropping database \"%s\"\n", db.Database)
	_, err = adminDB.Exec("drop database " + db.Database)
	return err
}

func LoadLogs(db pg.ConnSpec) error {
	c, err := db.Open()
	if err != nil {
		return err
	}
	ldr := loader.New(c, Sources(), CrawlSchema(),
		data.Crawl.StringMap("game-type-prefixes"))
	fmt.Println("Loading logs...")
	return ldr.LoadCommit()
}

func CreateIndexes(db pg.ConnSpec) error {
	c, err := db.Open()
	if err != nil {
		return err
	}
	sch := CrawlSchema().Schema().Sort()
	for _, index := range sch.SqlSel(schema.SelIndexes) {
		fmt.Println("EXEC", index)
		if _, err = c.Exec(index); err != nil {
			return err
		}
	}
	return nil
}
