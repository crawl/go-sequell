package db

import (
	"bytes"
	"database/sql"
	"log"

	"github.com/crawl/go-sequell/crawl/db"
	"github.com/crawl/go-sequell/loader"
	"github.com/crawl/go-sequell/pg"
)

const updateRowCount = 10000

func FixGodEcumenical(dbc pg.ConnSpec) error {
	c, err := dbc.Open()
	if err != nil {
		return err
	}

	sch := CrawlSchema()

	lookup := loader.NewTableLookup(sch.LookupTable("noun"), updateRowCount)
	for _, prefix := range sch.TableVariantPrefixes {
		if table := sch.Table("milestone"); table != nil {
			if err := fixTableGodEcumenical(c, sch, lookup, prefix+table.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func fixTableGodEcumenical(c pg.DB, sch *db.CrawlSchema, lookup *loader.TableLookup, table string) error {
	log.Println("Updating god.ecumenical rows in", table)
	rows, err := findGodEcumenicalRows(c, table)
	if err != nil {
		return err
	}
	defer rows.Close()

	type GodEcumenicalRow struct {
		id  int64
		god string
	}

	var godRows = make([]GodEcumenicalRow, updateRowCount)
	godRowCount := 0

	buildGodFixQuery := func(nrows int) string {
		buf := bytes.Buffer{}
		buf.WriteString(`update ` + table + ` as t set noun_id = c.nid from (values `)
		binder := pg.NewBinder()
		for i := 0; i < nrows; i++ {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("(" + binder.Next() + "::int,")
			buf.WriteString(binder.Next() + "::int)")
		}
		buf.WriteString(") as c (id, nid) where t.id = c.id")
		return buf.String()
	}

	updateGods := func() error {
		if godRowCount == 0 {
			return nil
		}
		for i := 0; i < godRowCount; i++ {
			lookup.AddLookup(godRows[i].god, nil)
		}
		tx, err := c.Begin()
		if err != nil {
			return err
		}
		if err = lookup.ResolveAll(tx); err != nil {
			tx.Rollback()
			return err
		}

		query := buildGodFixQuery(godRowCount)
		binds := make([]interface{}, godRowCount*2)
		for i := 0; i < godRowCount; i++ {
			binds[i*2] = godRows[i].id
			binds[i*2+1], err = lookup.Id(godRows[i].god)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
		log.Println("Updating", godRowCount, "rows of", table, "with fixed god nouns")
		_, err = tx.Exec(query, binds...)
		if err != nil {
			tx.Rollback()
			return err
		}
		godRowCount = 0
		return tx.Commit()
	}

	addGodRow := func(godRow GodEcumenicalRow) error {
		godRows[godRowCount] = godRow
		godRowCount++
		if godRowCount == len(godRows) {
			return updateGods()
		}
		return nil
	}

	var godRow GodEcumenicalRow
	for rows.Next() {
		if err := rows.Scan(&godRow.id, &godRow.god); err != nil {
			return err
		}
		addGodRow(godRow)
	}
	if err = updateGods(); err != nil {
		return err
	}
	return rows.Err()
}

func findGodEcumenicalRows(c pg.DB, table string) (*sql.Rows, error) {
	return c.Query(`select t.id, g.god from ` + table + ` as t
                    inner join l_god g on t.god_id = g.id
                    inner join l_verb v on t.verb_id = v.id
                    where v.verb = 'god.ecumenical'`)
}
