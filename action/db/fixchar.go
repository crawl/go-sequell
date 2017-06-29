package db

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/crawl/go-sequell/crawl/data"
	"github.com/crawl/go-sequell/crawl/db"
	"github.com/crawl/go-sequell/crawl/player"
	"github.com/crawl/go-sequell/loader"
	"github.com/crawl/go-sequell/pg"
	"github.com/crawl/go-sequell/stringnorm"
)

// FixCharFields fixes the "char" field in the db if the species and class
// fields disagree with the species and class mentioned in the char field (this
// was a Crawl bug that was subsequently fixed).
func FixCharFields(dbc pg.ConnSpec) error {
	c, err := dbc.Open()
	if err != nil {
		return err
	}

	sch := CrawlSchema()

	norm := player.StockCharNormalizer(data.CrawlData().YAML)
	for _, table := range sch.PrimaryTableNames() {
		rows, err := findMismatchedCharRows(c, norm, table)
		if err != nil {
			return err
		}
		defer rows.Close()
		if err = updateMismatchedCharRows(c, norm, sch, table, rows); err != nil {
			return err
		}
	}
	return nil
}

func updateMismatchedCharRows(c pg.DB, norm *player.CharNormalizer, sch *db.CrawlSchema, table string, rows *sql.Rows) error {
	type MismatchRow struct {
		id                int64
		race, class, abbr string
	}

	const bufferSize = 500
	rowCount := 0
	mismatchedRows := make([]MismatchRow, bufferSize)

	charLookup := loader.NewTableLookup(sch.LookupTable("char"), bufferSize)

	buildUpdateQuery := func(nrows int) string {
		buf := bytes.Buffer{}
		buf.WriteString(
			`update ` + table + ` as t set charabbrev_id = c.cid from
				(values `)
		binder := pg.NewBinder()
		for i := 0; i < nrows; i++ {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("(" + binder.Next() + "::int, ")
			buf.WriteString(binder.Next() + "::int)")
		}
		buf.WriteString(
			`) as c (id, cid) where t.id = c.id`)
		return buf.String()
	}

	commitMismatchFixes := func() error {
		if rowCount == 0 {
			return nil
		}
		var err error
		for i := 0; i < rowCount; i++ {
			newAbbr :=
				norm.NormalizeChar(
					mismatchedRows[i].race,
					mismatchedRows[i].class,
					mismatchedRows[i].abbr)
			if newAbbr == mismatchedRows[i].abbr {
				return fmt.Errorf("abbr(%s,%s) = %s; wanted change",
					mismatchedRows[i].race, mismatchedRows[i].class,
					newAbbr)
			}
			mismatchedRows[i].abbr = newAbbr
			charLookup.AddLookup(newAbbr, nil)
		}
		tx, err := c.Begin()
		if err != nil {
			return err
		}
		if err = charLookup.ResolveQueued(tx); err != nil {
			tx.Rollback()
			return err
		}
		updateQuery := buildUpdateQuery(bufferSize)
		binds := make([]interface{}, len(mismatchedRows)*2)
		for i := 0; i < rowCount; i++ {
			binds[i*2] = mismatchedRows[i].id
			binds[i*2+1], err = charLookup.ID(mismatchedRows[i].abbr)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
		log.Println("Updating", rowCount, "rows of", table,
			"with corrected char abbreviations")
		_, err = tx.Exec(updateQuery, binds...)
		if err != nil {
			tx.Rollback()
			return err
		}
		rowCount = 0
		return tx.Commit()
	}

	var mrow MismatchRow
	for rows.Next() {
		if err := rows.Scan(&mrow.id, &mrow.race, &mrow.class, &mrow.abbr); err != nil {
			return err
		}
		log.Printf("Row: %d (%s, %s) = %s\n", mrow.id, mrow.race, mrow.class, mrow.abbr)
		mismatchedRows[rowCount] = mrow
		rowCount++
		if rowCount >= len(mismatchedRows) {
			commitMismatchFixes()
		}
	}
	if err := commitMismatchFixes(); err != nil {
		return err
	}
	return rows.Err()
}

func listPairs(smap stringnorm.MultiMapper) map[string]string {
	res := map[string]string{}
	for key, val := range smap {
		res[key] = val[0]
	}
	return res
}

func mismatchedCharQuery(table string, nspecies, nclasses int) string {
	buf := bytes.Buffer{}
	buf.WriteString(
		`select t.id, r.crace, c.cls, ch.charabbrev
		   from ` + table + ` as t,
				l_crace as r,
				l_cls as c,
				l_char as ch
		   where t.crace_id = r.id and t.cls_id = c.id
			 and t.charabbrev_id = ch.id
			 and (`)
	binder := pg.NewBinder()
	for i := 0; i < nspecies; i++ {
		if binder.NotFirst() {
			buf.WriteString(" or ")
		}
		buf.WriteString("(r.crace = " + binder.Next())
		buf.WriteString(" and lower(substr(ch.charabbrev, 1, 2)) != " +
			binder.Next() + ")")
	}
	for i := 0; i < nclasses; i++ {
		buf.WriteString(" or ")
		buf.WriteString("(c.cls = " + binder.Next())
		buf.WriteString(" and lower(substr(ch.charabbrev, 3, 2)) != " +
			binder.Next() + ")")
	}
	buf.WriteString(")")
	return buf.String()
}

func pairwiseBinds(maps ...map[string]string) []interface{} {
	size := 0
	for _, m := range maps {
		size += len(m)
	}
	binds := make([]interface{}, size*2)
	i := 0
	for _, m := range maps {
		for k, v := range m {
			binds[i] = k
			binds[i+1] = strings.ToLower(v)
			i += 2
		}
	}
	return binds
}

func findMismatchedCharRows(c pg.DB, norm *player.CharNormalizer, table string) (*sql.Rows, error) {
	speciesNamesAbbrevs := listPairs(norm.SpeciesNameAbbrMap)
	classNamesAbbrevs := listPairs(norm.ClassNameAbbrMap)
	query := mismatchedCharQuery(table, len(speciesNamesAbbrevs), len(classNamesAbbrevs))
	binds := pairwiseBinds(speciesNamesAbbrevs, classNamesAbbrevs)
	log.Println("Querying", table, "for bad char abbreviations")
	return c.Query(query, binds...)
}
