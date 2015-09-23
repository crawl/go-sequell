package db

import (
	"bytes"
	dsql "database/sql"
	"fmt"

	"github.com/crawl/go-sequell/crawl/data"
	"github.com/crawl/go-sequell/crawl/db"
	"github.com/crawl/go-sequell/crawl/xlogtools"
	"github.com/crawl/go-sequell/loader"
	"github.com/crawl/go-sequell/pg"
	"github.com/crawl/go-sequell/sql"
	"github.com/crawl/go-sequell/stringnorm"
)

type FieldFixer struct {
	c                pg.DB
	schema           *db.CrawlSchema
	fieldGen         *xlogtools.FieldGen
	StringTransforms [][]string
	RegexpTransforms [][]string
}

func FixField(dbc pg.ConnSpec, field string) error {
	c, err := dbc.Open()
	if err != nil {
		return err
	}

	sch := CrawlSchema()
	gen, err := findFieldGen(field)
	if err != nil {
		return err
	}

	f := FieldFixer{c: c, schema: sch, fieldGen: gen}
	if err := f.init(); err != nil {
		return err
	}
	return f.FixFields()
}

func findFieldGen(field string) (*xlogtools.FieldGen, error) {
	norm, err := xlogtools.BuildNormalizer(data.Crawl)
	if err != nil {
		return nil, err
	}
	for _, gen := range norm.FieldGens {
		if gen.TargetField == field {
			return gen, nil
		}
	}
	return nil, fmt.Errorf("no field transform for %s", field)
}

func (f *FieldFixer) init() error {
	for _, norm := range stringnorm.NormList(f.fieldGen.Transforms) {
		switch xnorm := norm.(type) {
		case *stringnorm.ExactReplacer:
			f.StringTransforms =
				append(f.StringTransforms, []string{xnorm.Before, xnorm.After})
		case *stringnorm.RegexpNormalizer:
			f.RegexpTransforms =
				append(f.RegexpTransforms, []string{xnorm.Regexp.String(), xnorm.Replacement})
		default:
			return fmt.Errorf("Unexpected normalizer type: %T", xnorm)
		}
	}
	return nil
}

func (f *FieldFixer) TargetFieldName() string {
	return f.fieldGen.TargetField
}

func (f *FieldFixer) FixFields() error {
	for _, t := range f.schema.PrefixedTablesWithField(f.TargetFieldName()) {
		query, binds, err := f.mismatchedFieldQuery(t)
		if err != nil {
			return err
		}
		fmt.Println("Querying", t.Name, "for", f.TargetFieldName(), "needing fixups")

		rows, err := f.c.Query(query, binds...)
		if err != nil {
			return err
		}
		defer rows.Close()
		if err := f.updateBrokenFields(t, rows); err != nil {
			return err
		}
	}
	return nil
}

func (f *FieldFixer) updateBrokenFields(t *db.CrawlTable, rows *dsql.Rows) error {
	fixMap := map[string]string{}
	norm := f.fieldGen.Transforms
	const flushThreshold = 5000

	lookup := loader.NewTableLookup(f.schema.LookupTable(f.TargetFieldName()), flushThreshold)
	flush := func() error {
		if len(fixMap) == 0 {
			return nil
		}
		for _, newVal := range fixMap {
			lookup.AddLookup(newVal, nil)
		}
		txn, err := f.c.Begin()
		if err != nil {
			return err
		}
		if err = lookup.ResolveAll(txn); err != nil {
			txn.Rollback()
			return err
		}
		idmap := make(map[string]int, len(fixMap))
		fmt.Println("Updating", f.TargetFieldName(), "as", fixMap)
		for start, finish := range fixMap {
			delete(fixMap, start)
			id, err := lookup.ID(finish)
			if err != nil {
				txn.Rollback()
				return err
			}
			idmap[start] = id
		}
		updateQuery := f.fieldUpdateQuery(t, idmap)
		updateBinds := f.fieldUpdateBinds(idmap)
		_, err = txn.Exec(updateQuery, updateBinds...)
		if err != nil {
			txn.Rollback()
			return err
		}
		return txn.Commit()
	}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return err
		}

		if _, ok := fixMap[value]; ok {
			continue
		}

		res, err := norm.Normalize(value)
		if err != nil {
			return err
		}
		if res == value {
			continue
		}
		fixMap[value] = res
		if len(fixMap) > flushThreshold {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	if err := flush(); err != nil {
		return err
	}
	return rows.Err()
}

func (f *FieldFixer) fieldUpdateQuery(t *db.CrawlTable, valids map[string]int) string {
	field := t.FindField(f.TargetFieldName())
	query := bytes.Buffer{}
	query.WriteString(
		"update " + t.Name + " as t set " + field.RefName() + " = c.id")
	query.WriteString(" from " + field.ForeignKeyTable + " as l, ")
	query.WriteString("(values ")
	binder := pg.NewBinder()
	for _ = range valids {
		if binder.NotFirst() {
			query.WriteString(",")
		}
		query.WriteString("(")
		query.WriteString(binder.Next())
		query.WriteString(",")
		query.WriteString(binder.Next())
		query.WriteString("::int)")
	}
	query.WriteString(") as c (value, id)")
	query.WriteString(" where t." + field.RefName() + " = l.id")
	query.WriteString(" and l." + field.SqlName + " = c.value")
	return query.String()
}

func (f *FieldFixer) fieldUpdateBinds(valids map[string]int) []interface{} {
	res := make([]interface{}, 0, len(valids)*2)
	for key, id := range valids {
		res = append(res, key, id)
	}
	return res
}

func (f *FieldFixer) mismatchedFieldQuery(t *db.CrawlTable) (string, []interface{}, error) {
	query := sql.Select{}
	baseTable := query.AddNamedTable(t.Name)
	selectTable := baseTable
	field := t.FindField(f.TargetFieldName())
	if field.ForeignKeyLookup {
		lookupTable := &sql.Table{Name: field.ForeignKeyTable}
		joinTable := query.AddJoinTable(
			lookupTable,
			baseTable.QualifiedName(field.RefName())+" = "+
				lookupTable.QualifiedName("id"),
			"join")
		selectTable = joinTable
	}
	query.AddColumnExpr(selectTable.QualifiedName(field.SqlName))
	return query.SQL() + " where " + f.whereClause(field, selectTable.Table), f.queryBinds(), nil
}

func (f *FieldFixer) whereClause(field *db.Field, table *sql.Table) string {
	buf := bytes.Buffer{}
	fieldExpr := table.QualifiedName(field.SqlName)
	binder := pg.NewBinder()
	for _ = range f.StringTransforms {
		if binder.NotFirst() {
			buf.WriteString(" or ")
		}
		buf.WriteString(fieldExpr)
		buf.WriteString(" = ")
		buf.WriteString(binder.Next())
	}
	for _ = range f.RegexpTransforms {
		if binder.NotFirst() {
			buf.WriteString(" or ")
		}
		buf.WriteString(fieldExpr)
		buf.WriteString(" ~ ")
		buf.WriteString(binder.Next())
	}
	return buf.String()
}

func (f *FieldFixer) queryBinds() []interface{} {
	res := make([]interface{}, 0, len(f.StringTransforms)+len(f.RegexpTransforms))
	for _, s := range f.StringTransforms {
		res = append(res, s[0])
	}
	for _, r := range f.RegexpTransforms {
		res = append(res, r[0])
	}
	return res
}
