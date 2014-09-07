package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/greensnark/go-sequell/action"
	"github.com/greensnark/go-sequell/pg"
)

func main() {
	app := cli.NewApp()
	app.Name = "seqdb"
	app.Usage = "Sequell db operations"
	app.Version = "1.0.0"
	app.Action = func(c *cli.Context) {
		cli.ShowAppHelp(c)
	}
	defineFlags(app)
	defineCommands(app)
	app.Run(os.Args)
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
	}
}

func reportError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

func fatal(msg string) {
	fmt.Println(os.Stderr, msg)
	os.Exit(1)
}

func defineCommands(app *cli.App) {
	dbSpec := func(c *cli.Context) pg.ConnSpec {
		return pg.ConnSpec{
			Database: c.String("db"),
			User:     c.String("user"),
			Password: c.String("password"),
		}
	}
	app.Commands = []cli.Command{
		{
			Name:  "schema",
			Usage: "print the Sequell schema",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "no-index",
					Usage: "skip create index statements",
				},
				cli.BoolFlag{
					Name:  "drop-index",
					Usage: "drop index statements only",
				},
				cli.BoolFlag{
					Name:  "create-index",
					Usage: "create index statements only",
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
				action.PrintSchema(noIndex, dropIndex, createIndex)
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
				action.CheckDBSchema(dbSpec(c), c.Bool("upgrade"))
			},
		},
		{
			Name:  "createdb",
			Usage: "create the Sequell database",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "admin",
					Usage: "Postgres admin user (optional; to create schema)",
				},
				cli.StringFlag{
					Name:  "adminpassword",
					Usage: "Postgres admin user's password (optional; to create schema)",
				},
				cli.StringFlag{
					Name:  "admindb",
					Value: "postgres",
					Usage: "Postgres admin db",
				},
			},
			Action: func(c *cli.Context) {
				adminSpec := pg.ConnSpec{
					Database: c.String("admindb"),
					User:     c.String("admin"),
					Password: c.String("adminpassword"),
				}
				reportError(action.CreateDB(adminSpec, dbSpec(c)))
			},
		},
	}
}
