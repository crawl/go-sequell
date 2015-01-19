package main

import (
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/greensnark/go-sequell/action"
	"github.com/greensnark/go-sequell/action/db"
	"github.com/greensnark/go-sequell/pg"
)

var Error error

func main() {
	app := cli.NewApp()
	app.Name = "seqdb"
	app.Usage = "Sequell db ops"
	app.Version = "1.1.0"
	app.Action = func(c *cli.Context) {
		cli.ShowAppHelp(c)
	}
	defineFlags(app)
	defineCommands(app)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	app.Run(os.Args)
	if Error != nil {
		os.Exit(1)
	}
}

func defineFlags(app *cli.App) {
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "db",
			Value:  "sequell",
			Usage:  "Sequell database name",
			EnvVar: "SEQUELL_DBNAME",
		},
		cli.StringFlag{
			Name:   "user",
			Value:  "sequell",
			Usage:  "Sequell database user",
			EnvVar: "SEQUELL_DBUSER",
		},
		cli.StringFlag{
			Name:   "password",
			Value:  "sequell",
			Usage:  "Sequell database user password",
			EnvVar: "SEQUELL_DBPASS",
		},
		cli.StringFlag{
			Name:   "host",
			Value:  "localhost",
			Usage:  "Sequell postgres database host",
			EnvVar: "SEQUELL_DBHOST",
		},
		cli.IntFlag{
			Name:   "port",
			Value:  0,
			Usage:  "Sequell postgres database port",
			EnvVar: "SEQUELL_DBPORT",
		},
	}
}

func reportError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		Error = err
	}
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func adminFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "admin",
			Usage: "Postgres admin user (optional)",
		},
		cli.StringFlag{
			Name:  "adminpassword",
			Usage: "Postgres admin user's password (optional)",
		},
		cli.StringFlag{
			Name:  "admindb",
			Value: "postgres",
			Usage: "Postgres admin db",
		},
	}
}

func dropFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "force",
			Usage: "actually drop the database",
		},
		cli.BoolFlag{
			Name:  "terminate",
			Usage: "terminate other sessions connected to the database",
		},
	}
}

func adminDBSpec(c *cli.Context) pg.ConnSpec {
	return pg.ConnSpec{
		Database: c.String("admindb"),
		User:     c.String("admin"),
		Password: c.String("adminpassword"),
		Host:     c.GlobalString("host"),
		Port:     c.GlobalInt("port"),
	}
}

func defineCommands(app *cli.App) {
	dbSpec := func(c *cli.Context) pg.ConnSpec {
		return pg.ConnSpec{
			Database: c.GlobalString("db"),
			User:     c.GlobalString("user"),
			Password: c.GlobalString("password"),
			Host:     c.GlobalString("host"),
			Port:     c.GlobalInt("port"),
		}
	}
	app.Commands = []cli.Command{
		{
			Name:  "fetch",
			Usage: "download logs from all sources",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "only-live",
					Usage: "fetch only logs that are believed to be live",
				},
			},
			Action: func(c *cli.Context) {
				reportError(action.DownloadLogs(c.Bool("only-live")))
			},
		},
		{
			Name:  "load",
			Usage: "load all outstanding data in the logs to the db",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "force-source-dir",
					Usage: "Forces the loader to use the files in the directory specified, associating them with the appropriate servers (handy to load test data)",
				},
			},
			Action: func(c *cli.Context) {
				reportError(db.LoadLogs(dbSpec(c), c.String("force-source-dir")))
			},
		},
		{
			Name:  "isync",
			Usage: "load all data, then run an interactive process that accepts commands to \"fetch\" on stdin, automatically loading logs that are updated",
			Action: func(c *cli.Context) {
				reportError(action.Isync(dbSpec(c)))
			},
		},
		{
			Name:  "schema",
			Usage: "print the Sequell schema",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "no-index",
					Usage: "table drop+create DDL only; no indexes and constraints",
				},
				cli.BoolFlag{
					Name:  "drop-index",
					Usage: "DDL to drop indexes and constraints only; no tables",
				},
				cli.BoolFlag{
					Name:  "create-index",
					Usage: "DDL to create indexes and constraints only; no tables",
				},
			},
			Action: func(c *cli.Context) {
				noIndex := c.Bool("no-index")
				dropIndex := c.Bool("drop-index")
				createIndex := c.Bool("create-index")
				if noIndex && (dropIndex || createIndex) {
					fatal("--no-index cannot be combined with --drop-index or --create-index")
				}
				if dropIndex && createIndex {
					fatal("--drop-index cannot be combined with --create-index")
				}
				db.PrintSchema(noIndex, dropIndex, createIndex)
			},
		},
		{
			Name:  "dumpschema",
			Usage: "dump the schema currently in the db",
			Action: func(c *cli.Context) {
				db.DumpSchema(dbSpec(c))
			},
		},
		{
			Name:      "checkdb",
			ShortName: "c",
			Usage:     "check the DB schema for correctness",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "upgrade",
					Usage: "apply any changes to the DB",
				},
			},
			Action: func(c *cli.Context) {
				reportError(db.CheckDBSchema(dbSpec(c), c.Bool("upgrade")))
			},
		},
		{
			Name:  "newdb",
			Usage: "create the Sequell database and initialize it",
			Flags: adminFlags(),
			Action: func(c *cli.Context) {
				if err := db.CreateDB(adminDBSpec(c), dbSpec(c)); err != nil {
					reportError(err)
					return
				}
				reportError(db.CreateDBSchema(dbSpec(c)))
			},
		},
		{
			Name:  "dropdb",
			Usage: "drop the Sequell database (must use --force)",
			Flags: append(adminFlags(), dropFlags()...),
			Action: func(c *cli.Context) {
				reportError(
					db.DropDB(adminDBSpec(c), dbSpec(c), c.Bool("force"),
						c.Bool("terminate")))
			},
		},
		{
			Name:  "resetdb",
			Usage: "drop and recreate the Sequell database (must use --force), => dropdb + newdb",
			Flags: append(adminFlags(), dropFlags()...),
			Action: func(c *cli.Context) {
				force := c.Bool("force")
				reportError(
					db.DropDB(adminDBSpec(c), dbSpec(c), force,
						c.Bool("terminate")))
				if force {
					reportError(
						db.CreateDB(adminDBSpec(c), dbSpec(c)))
					reportError(db.CreateDBSchema(dbSpec(c)))
				}
			},
		},
		{
			Name:  "createdb",
			Usage: "create the Sequell database (empty)",
			Flags: adminFlags(),
			Action: func(c *cli.Context) {
				reportError(db.CreateDB(adminDBSpec(c), dbSpec(c)))
			},
		},
		{
			Name:  "create-tables",
			Usage: "create tables in the Sequell database",
			Action: func(c *cli.Context) {
				reportError(db.CreateDBSchema(dbSpec(c)))
			},
		},
		{
			Name:  "create-indexes",
			Usage: "create indexes (use after loading)",
			Action: func(c *cli.Context) {
				reportError(db.CreateIndexes(dbSpec(c)))
			},
		},
		{
			Name:  "rm-file",
			Usage: "deletes rows inserted from the specified file(s)",
			Action: func(c *cli.Context) {
				reportError(db.DeleteFileRows(dbSpec(c), c.Args()))
			},
		},
		{
			Name:  "export-tv",
			Usage: "export ntv data (writes to stdout)",
			Action: func(c *cli.Context) {
				reportError(db.ExportTV(dbSpec(c)))
			},
		},
		{
			Name:  "import-tv",
			Usage: "import ntv data (reads from stdin)",
			Action: func(c *cli.Context) {
				reportError(db.ImportTV(dbSpec(c)))
			},
		},
		{
			Name:  "vrenum",
			Usage: "recomputes version numbers for l_version, l_cversion and l_vlong. Use this to update these tables if/when the version number algorithm changes.",
			Action: func(c *cli.Context) {
				reportError(db.RenumberVersions(dbSpec(c)))
			},
		},
		{
			Name:  "fix-char",
			Usage: "fix incorrect `char` fields using crace and cls",
			Action: func(c *cli.Context) {
				reportError(db.FixCharFields(dbSpec(c)))
			},
		},
		{
			Name:  "fix-field",
			Usage: "fix incorrect field",
			Action: func(c *cli.Context) {
				if len(c.Args()) <= 0 {
					reportError(fmt.Errorf("field to fix not specified"))
					return
				}
				reportError(db.FixField(dbSpec(c), c.Args()[0]))
			},
		},
	}
}