package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/crawl/go-sequell/action"
	"github.com/crawl/go-sequell/action/db"
	"github.com/crawl/go-sequell/pg"
	"github.com/crawl/go-sequell/text"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	cmd     = "seqdb"
	version = "1.1.0"
)

var cmdError error

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	app := &cobra.Command{
		Use:   cmd,
		Short: cmd + " manages Sequell's game database",
		PreRun: func(c *cobra.Command, args []string) {
			if logp := c.Flag("log"); logp != nil {
				if logPath := logp.Value.String(); logPath != "" {
					if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
						fmt.Fprintln(os.Stderr, "mkdir", logPath, "failed:", err)
						return
					}
					log.SetOutput(&lumberjack.Logger{
						Filename:   logPath,
						MaxSize:    10,
						MaxBackups: 10,
						MaxAge:     15,
					})
				}
			}
		},
	}
	defineAppFlags(app)
	defineCommands(app)
	app.Execute()
}

func defineAppFlags(app *cobra.Command) {
	f := app.PersistentFlags()
	f.String("log", os.Getenv("SEQUELL_LOG"), "Sequell log file path")
	f.String("db", text.FirstNotEmpty(os.Getenv("SEQUELL_DBNAME"), "sequell"), "Sequell database name")
	f.String("user", text.FirstNotEmpty(os.Getenv("SEQUELL_DBUSER"), "sequell"), "Sequell database user")
	f.String("password", text.FirstNotEmpty(os.Getenv("SEQUELL_DBPASS"), "sequell"), "Sequell database password")
	f.String("host", text.FirstNotEmpty(os.Getenv("SEQUELL_DBHOST"), "localhost"), "Sequell postgres database host")
	f.Int("port", text.ParseInt("SEQUELL_DBPORT", 0), "Sequell postgres database port")
}

func reportError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		cmdError = err
	}
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func adminFlags(f *pflag.FlagSet) {
	f.String("admin", "", "Postgres admin user (optional)")
	f.String("adminpassword", "", "Postgres admin user's password (optional)")
	f.String("admindb", "postgres", "Postgres admin db")
}

func dropFlags(f *pflag.FlagSet) {
	f.Bool("force", false, "actually drop the database")
	f.Bool("terminate", false, "terminate other sessions connected to the database")
}

func adminDBSpec(c *cobra.Command) pg.ConnSpec {
	return pg.ConnSpec{
		Database: stringFlag(c, "admindb"),
		User:     stringFlag(c, "admin"),
		Password: stringFlag(c, "adminpassword"),
		Host:     stringFlag(c, "host"),
		Port:     intFlag(c, "port"),
	}
}

func setFlags(flagSetter func(*pflag.FlagSet), cmd *cobra.Command) *cobra.Command {
	flagSetter(cmd.Flags())
	return cmd
}

func andFlags(flagSetters ...func(*pflag.FlagSet)) func(*pflag.FlagSet) {
	return func(f *pflag.FlagSet) {
		for _, setter := range flagSetters {
			setter(f)
		}
	}
}

func boolFlag(cmd *cobra.Command, name string) bool {
	val, err := cmd.Flags().GetBool(name)
	if err != nil {
		fatal("bad boolean value for" + name + ": " + err.Error())
	}
	return val
}

func stringFlag(cmd *cobra.Command, name string) string {
	val, err := cmd.Flags().GetString(name)
	if err != nil {
		fatal("bad string value for " + name + ": " + err.Error())
	}
	return val
}

func intFlag(cmd *cobra.Command, name string) int {
	val, err := cmd.Flags().GetInt(name)
	if err != nil {
		fatal("bad int value for " + name + ": " + err.Error())
	}
	return val
}

func defineCommands(app *cobra.Command) {
	app.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show " + cmd + " version",
		Run: func(*cobra.Command, []string) {
			fmt.Println(cmd, version)
		},
	})

	dbSpec := func(c *cobra.Command) pg.ConnSpec {
		return pg.ConnSpec{
			Database: stringFlag(c, "db"),
			User:     stringFlag(c, "user"),
			Password: stringFlag(c, "password"),
			Host:     stringFlag(c, "host"),
			Port:     intFlag(c, "port"),
		}
	}

	app.AddCommand(setFlags(func(f *pflag.FlagSet) {
		f.Bool("only-live", false, "fetch only logs that are live")
	}, &cobra.Command{
		Use:   "fetch",
		Short: "download logs from all sources",
		Run: func(c *cobra.Command, args []string) {
			reportError(action.DownloadLogs(boolFlag(c, "only-live"), args))
		},
	}))

	app.AddCommand(setFlags(func(f *pflag.FlagSet) {
		f.String("force-source-dir", "", "Forces the loader to use the files in the directory specified, associating them with appropriate servers (for test data)")
	}, &cobra.Command{
		Use:   "load",
		Short: "load all outstanding data in the logs to the db",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.LoadLogs(dbSpec(c), stringFlag(c, "force-source-dir")))
		},
	}))

	app.AddCommand(&cobra.Command{
		Use:   "isync",
		Short: "load all data, then run an interactive process that accepts commands to \"fetch\" on stdin, automatically loading logs that are updated",
		Run: func(c *cobra.Command, args []string) {
			reportError(action.Isync(dbSpec(c)))
		},
	})
	app.AddCommand(setFlags(func(f *pflag.FlagSet) {
		f.Bool("no-index", false, "table drop+create DDL only; no indexes and constraints")
		f.Bool("drop-index", false, "DDL to drop indexes and constraints only; no tables")
		f.Bool("create-index", false, "DDL to create indexes and constraints only; no tables")
	}, &cobra.Command{
		Use:   "schema",
		Short: "print the Sequell schema",
		Run: func(c *cobra.Command, args []string) {
			noIndex := boolFlag(c, "no-index")
			dropIndex := boolFlag(c, "drop-index")
			createIndex := boolFlag(c, "create-index")
			if noIndex && (dropIndex || createIndex) {
				fatal("--no-index cannot be combined with --drop-index or --create-index")
			}
			if dropIndex && createIndex {
				fatal("--drop-index cannot be combined with --create-index")
			}
			db.PrintSchema(noIndex, dropIndex, createIndex)
		},
	}))

	app.AddCommand(&cobra.Command{
		Use:   "dumpschema",
		Short: "dump the schema currently in the db",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.DumpSchema(dbSpec(c)))
		},
	})
	app.AddCommand(setFlags(func(f *pflag.FlagSet) {
		f.Bool("upgrade", false, "apply any changes to the DB (not implemented)")
	}, &cobra.Command{
		Use:   "checkdb",
		Short: "check the DB schema for correctness",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.CheckDBSchema(dbSpec(c), boolFlag(c, "upgrade")))
		},
	}))
	app.AddCommand(setFlags(adminFlags, &cobra.Command{
		Use:   "newdb",
		Short: "create the Sequell database and initialize it",
		Run: func(c *cobra.Command, args []string) {
			if err := db.CreateDB(adminDBSpec(c), dbSpec(c)); err != nil {
				reportError(err)
				return
			}
			reportError(db.CreateDBSchema(dbSpec(c)))
		},
	}))
	app.AddCommand(setFlags(andFlags(adminFlags, dropFlags), &cobra.Command{
		Use:   "dropdb",
		Short: "drop the Sequell database (must use --force)",
		Run: func(c *cobra.Command, args []string) {
			reportError(
				db.DropDB(adminDBSpec(c), dbSpec(c), boolFlag(c, "force"),
					boolFlag(c, "terminate")))
		},
	}))
	app.AddCommand(setFlags(andFlags(adminFlags, dropFlags), &cobra.Command{
		Use:   "resetdb",
		Short: "drop and recreate the Sequell database (must use --force), => dropdb + newdb",
		Run: func(c *cobra.Command, args []string) {
			force := boolFlag(c, "force")
			reportError(
				db.DropDB(adminDBSpec(c), dbSpec(c), force,
					boolFlag(c, "terminate")))
			if force {
				reportError(
					db.CreateDB(adminDBSpec(c), dbSpec(c)))
				reportError(db.CreateDBSchema(dbSpec(c)))
			}
		},
	}))
	app.AddCommand(setFlags(adminFlags, &cobra.Command{
		Use:   "createdb",
		Short: "create the Sequell database (empty)",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.CreateDB(adminDBSpec(c), dbSpec(c)))
		},
	}))
	app.AddCommand(&cobra.Command{
		Use:   "create-tables",
		Short: "create tables in the Sequell database",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.CreateDBSchema(dbSpec(c)))
		},
	})
	app.AddCommand(&cobra.Command{
		Use:   "create-indexes",
		Short: "create indexes (use after loading)",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.CreateIndexes(dbSpec(c)))
		},
	})
	app.AddCommand(&cobra.Command{
		Use:   "ls-files",
		Short: "lists all files known to Sequell",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.ListFiles(dbSpec(c)))
		},
	})
	app.AddCommand(&cobra.Command{
		Use:   "rm-file",
		Short: "deletes rows inserted from the specified file(s)",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.DeleteFileRows(dbSpec(c), args))
		},
	})
	app.AddCommand(&cobra.Command{
		Use:   "sources",
		Short: "show all remote source URLs",
		Run: func(c *cobra.Command, args []string) {
			reportError(action.ShowSourceURLs())
		},
	})
	app.AddCommand(&cobra.Command{
		Use:   "export-tv",
		Short: "export ntv data (writes to stdout)",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.ExportTV(dbSpec(c)))
		},
	})
	app.AddCommand(&cobra.Command{
		Use:   "import-tv",
		Short: "import ntv data (reads from stdin)",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.ImportTV(dbSpec(c)))
		},
	})
	app.AddCommand(&cobra.Command{
		Use:   "vrenum",
		Short: "recomputes version numbers for l_version, l_cversion and l_vlong. Use this to update these tables if/when the version number algorithm changes.",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.RenumberVersions(dbSpec(c)))
		},
	})
	app.AddCommand(&cobra.Command{
		Use:   "fix-char",
		Short: "fix incorrect `char` fields using crace and cls",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.FixCharFields(dbSpec(c)))
		},
	})
	app.AddCommand(&cobra.Command{
		Use:   "fix-field",
		Short: "fix incorrect field",
		Run: func(c *cobra.Command, args []string) {
			if len(args) <= 0 {
				reportError(fmt.Errorf("field to fix not specified"))
				return
			}
			reportError(db.FixField(dbSpec(c), args[0]))
		},
	})
	app.AddCommand(&cobra.Command{
		Use:   "fix-god-ecumenical",
		Short: "fix nouns for god.ecumenical milestones",
		Run: func(c *cobra.Command, args []string) {
			reportError(db.FixGodEcumenical(dbSpec(c)))
		},
	})
	app.AddCommand(&cobra.Command{
		Use:   "xlog-link",
		Short: "link old remote.* to new URL-based paths",
		Run: func(c *cobra.Command, args []string) {
			reportError(action.LinkLogs())
		},
	})
}
