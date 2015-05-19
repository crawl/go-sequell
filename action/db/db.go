package db

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/crawl/go-sequell/action"
	"github.com/crawl/go-sequell/crawl/data"
	cdb "github.com/crawl/go-sequell/crawl/db"
	"github.com/crawl/go-sequell/crawl/xlogtools"
	"github.com/crawl/go-sequell/ectx"
	"github.com/crawl/go-sequell/loader"
	"github.com/crawl/go-sequell/pg"
	"github.com/crawl/go-sequell/schema"
	"github.com/crawl/go-sequell/sources"
)

var DBExtensions = []string{"citext", "orafce"}

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
	s.Sort().Write(schema.SelTablesIndexesConstraints, os.Stdout)
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
		log.Printf("Creating database \"%s\"\n", db.Database)
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
		log.Printf("Creating user \"%s\"\n", dbspec.User)
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
	for _, ext := range DBExtensions {
		extExists, err := c.ExtensionExists(ext)
		if err != nil {
			return err
		}
		if !extExists {
			log.Printf("Creating extension \"%s\"\n", ext)
			if err = c.CreateExtension(ext); err != nil {
				return err
			}
		}
	}
	return nil
}

func PrintSchema(skipIndexes, dropIndexes, createIndexes bool) {
	s := CrawlSchema().Schema()
	sel := schema.SelTablesIndexesConstraints
	if skipIndexes {
		sel = schema.SelTables
	}
	if dropIndexes {
		sel = schema.SelDropIndexesConstraints
	}
	if createIndexes {
		sel = schema.SelIndexesConstraints
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
		log.Println("Schema is up-to-date.")
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
	log.Printf("Creating tables in database \"%s\"\n", db.Database)
	for _, sql := range s.SqlSel(schema.SelTables) {
		if _, err = c.Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func DropDB(admin pg.ConnSpec, db pg.ConnSpec, force, terminate bool) error {
	if !force {
		return fmt.Errorf("Use --force to drop the database \"%s\"",
			db.Database)
	}
	adminDB, err := admin.Open()
	if err != nil {
		return err
	}

	if terminate {
		if err = TerminateConnections(adminDB, db.Database); err != nil {
			return err
		}
	}

	log.Printf("Dropping database \"%s\"\n", db.Database)
	_, err = adminDB.Exec("drop database " + db.Database)
	return err
}

func TerminateConnections(adminDB pg.DB, targetDB string) error {
	pids, err := adminDB.ActiveConnections(targetDB)
	if err != nil {
		return err
	}
	for _, pid := range pids {
		log.Println("Terminating backend", pid)
		if err = adminDB.TerminateConnection(pid); err != nil {
			return ectx.Err(fmt.Sprintf("[%d]", pid), err)
		}
	}
	return nil
}

func LoadLogs(db pg.ConnSpec, sourceDir string) error {
	c, err := db.Open()
	if err != nil {
		return err
	}
	defer c.Close()

	sources := Sources()
	if sourceDir != "" {
		if err = forceSourceDir(sources, sourceDir); err != nil {
			return err
		}
	}

	if err := action.DBLock.Lock(false); err != nil {
		return err
	}
	defer action.DBLock.Unlock()

	logNorm := xlogtools.MustBuildNormalizer(data.Crawl)
	ldr := loader.New(c, sources, CrawlSchema(), logNorm,
		data.Crawl.StringMap("game-type-prefixes"))

	if sourceDir != "" {
		log.Println("Loading logs from", sourceDir, "into", db.Database)
	} else {
		log.Println("Loading logs into", db.Database)
	}
	return ldr.LoadCommit()
}

func forceSourceDir(srv *sources.Servers, dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	// Zap all logs and milestones
	for _, server := range srv.Servers {
		server.Logfiles = nil
	}

	sourceMap := map[string][]*sources.XlogSrc{}
	for _, f := range files {
		filename := f.Name()
		if !xlogtools.IsXlogQualifiedName(filename) {
			continue
		}
		src, game, xtype := xlogtools.XlogServerType(filename)
		if xtype == xlogtools.Unknown {
			log.Printf("Ignoring %s: unknown type\n", filename)
			continue
		}

		server := srv.Server(src)
		if server == nil {
			log.Printf("Ignoring %s: can't find server %s\n", filename, src)
			continue
		}

		xl := sources.XlogSrc{
			Server:        server,
			Name:          filename,
			TargetPath:    path.Join(dir, filename),
			TargetRelPath: filename,
			Type:          xtype,
			Game:          game,
		}
		sourceMap[src] = append(sourceMap[src], &xl)
	}

	for src, xlogs := range sourceMap {
		server := srv.Server(src)
		server.Logfiles = xlogs
	}
	return nil
}

func CreateIndexes(db pg.ConnSpec) error {
	c, err := db.Open()
	if err != nil {
		return err
	}
	sch := CrawlSchema().Schema().Sort()
	for _, index := range sch.SqlSel(schema.SelIndexesConstraints) {
		log.Println("Exec:", index)
		if _, err = c.Exec(index); err != nil {
			log.Printf("Error creating index: %s\n", err)
		}
	}
	return nil
}

func ListFiles(db pg.ConnSpec) error {
	c, err := db.Open()
	if err != nil {
		return err
	}
	defer c.Close()

	rows, err := c.Query("select file from l_file")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var file string
		if err = rows.Scan(&file); err != nil {
			return err
		}
		fmt.Println(file)
	}
	return rows.Err()
}

func DeleteFileRows(db pg.ConnSpec, files []string) error {
	if len(files) == 0 {
		log.Println("No files specified.")
		return nil
	}
	c, err := db.Open()
	if err != nil {
		return err
	}
	defer c.Close()

	binds := func(nbind int) string {
		buf := bytes.Buffer{}
		for i := 0; i < nbind; i++ {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("$")
			buf.WriteString(strconv.Itoa(i + 1))
		}
		return buf.String()
	}

	fileArgs := func(s []string) []interface{} {
		res := make([]interface{}, len(s))
		for i, v := range s {
			res[i] = v
		}
		return res
	}

	fileStr := func(files []interface{}) string {
		buf := bytes.Buffer{}
		for i, f := range files {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(f.(string))
		}
		return buf.String()
	}

	ifiles := fileArgs(files)
	log.Printf("Deleting rows from %d files: %s\n",
		len(ifiles), fileStr(ifiles))
	query := "delete from l_file where file in (" + binds(len(files)) + ")"
	res, err := c.Exec(query, ifiles...)
	if err != nil {
		return err
	}
	if rowsAffected, err := res.RowsAffected(); err == nil {
		log.Println("Deleted", rowsAffected, "rows")
	}
	return nil
}
