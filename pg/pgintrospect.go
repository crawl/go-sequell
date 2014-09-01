package pg

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/greensnark/go-sequell/schema"
	"strconv"
)

func (p PgDB) IntrospectSchema() (*schema.Schema, error) {
	tables, err := p.IntrospectTableNames()
	if err != nil {
		return nil, ctx("IntrospectSchema", err)
	}

	tableSchemas, err := p.IntrospectTableSchemas(tables)
	if err != nil {
		return nil, ctx("IntrospectSchema", err)
	}

	schema := schema.Schema{
		Tables: tableSchemas,
	}
	return &schema, nil
}

func (p PgDB) IntrospectTableNames() ([]string, error) {
	rows, err := p.Query("select table_name from information_schema.tables where table_schema = 'public' and table_type = 'BASE TABLE'")
	if err != nil {
		return nil, err
	}

	tableNames := []string{}
	defer rows.Close()
	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			return nil, ctx("IntrospectTableNames", err)
		}
		tableNames = append(tableNames, tableName)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tableNames, nil
}

func (p PgDB) IntrospectTableSchemas(tables []string) (schemas []*schema.Table, err error) {
	schemas = make([]*schema.Table, len(tables))
	for i, t := range tables {
		schemas[i], err = p.IntrospectTable(t)
		if err != nil {
			return nil, err
		}
	}
	return schemas, nil
}

func (p PgDB) IntrospectTable(table string) (*schema.Table, error) {
	cols, err := p.IntrospectTableColumns(table)
	if err != nil {
		return nil, err
	}

	constraints, err := p.IntrospectTableConstraints(table, cols)
	if err != nil {
		return nil, err
	}

	indexes, err := p.IntrospectTableIndexes(table, cols)
	if err != nil {
		return nil, err
	}

	return &schema.Table{
		Name:        table,
		Columns:     cols,
		Indexes:     indexes,
		Constraints: constraints,
	}, nil
}

func (p PgDB) IntrospectTableColumns(table string) ([]*schema.Column, error) {
	rows, err := p.Query(`select column_name, column_default,
                                 data_type, udt_name, numeric_precision
                            from information_schema.columns
                           where table_schema = 'public' and table_name = $1
                        order by ordinal_position`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := []*schema.Column{}
	for rows.Next() {
		var colName, colDefault, dataType, userDefinedType sql.NullString
		var numericPrecision sql.NullInt64
		err = rows.Scan(&colName, &colDefault, &dataType, &userDefinedType, &numericPrecision)
		if err != nil {
			return nil, ctx("IntrospectTableColumns", err)
		}
		if dataType.String == "USER-DEFINED" {
			dataType = userDefinedType
		}
		cols = append(cols, p.ColumnDef(colName.String, colDefault.String, dataType.String, numericPrecision.Int64))
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return cols, nil
}

func (p PgDB) ColumnDef(name, defval, dataType string, precision int64) *schema.Column {
	dataType, defval = p.NormalizeType(dataType, defval, precision)
	return &schema.Column{
		Name:    name,
		SqlType: dataType,
		Default: defval,
	}
}

func (p PgDB) NormalizeType(sqlType, defval string, precision int64) (string, string) {
	if sqlType == "timestamp without time zone" {
		sqlType = "timestamp"
	}
	if sqlType == "integer" {
		sqlType = "int"
	}
	// Sequences are normalized to serial.
	if strings.Index(defval, "nextval(") != -1 && sqlType == "int" {
		sqlType = "serial"
		defval = ""
	}
	if sqlType == "numeric" && precision > 0 {
		sqlType = "numeric(" + strconv.FormatInt(precision, 10) + ")"
	}
	return sqlType, defval
}

func (p PgDB) IntrospectTableConstraints(table string, cols []*schema.Column) ([]schema.Constraint, error) {
	pkey, err := p.IntrospectTablePrimaryKey(table)
	if err != nil {
		return nil, err
	}

	constraints := []schema.Constraint{}
	if pkey != nil {
		constraints = append(constraints, pkey)
	}

	fkeys, err := p.IntrospectTableForeignKeys(table)
	if err != nil {
		return nil, err
	}
	constraints = append(constraints, fkeys...)

	err = p.IntrospectTableUniqueConstraints(table, cols)
	if err != nil {
		return nil, err
	}

	return constraints, nil
}

func (p PgDB) IntrospectTablePrimaryKey(table string) (schema.Constraint, error) {
	row := p.QueryRow(`select kcu.column_name
                              from information_schema.table_constraints as tc
                              join information_schema.key_column_usage as kcu
                                on tc.constraint_name = kcu.constraint_name
                             where tc.constraint_type = 'PRIMARY KEY'
                               and tc.table_name = $1`, table)
	var pkey string
	if err := row.Scan(&pkey); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, ctx("IntrospectTablePrimaryKey", err)
	}
	return schema.PrimaryKeyConstraint{Column: pkey}, nil
}

func (p PgDB) IntrospectTableForeignKeys(table string) ([]schema.Constraint, error) {
	rows, err := p.Query(`select kcu.column_name,
                                 ccu.table_name as foreign_table,
                                 ccu.column_name as foreign_column
                            from information_schema.table_constraints as tc
                            join information_schema.key_column_usage as kcu
                              on tc.constraint_name = kcu.constraint_name
                            join information_schema.constraint_column_usage as ccu
                              on tc.constraint_name = ccu.constraint_name
                           where tc.constraint_type = 'FOREIGN KEY'
                             and tc.table_name = $1`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	constraints := []schema.Constraint{}
	for rows.Next() {
		var col, foreignTable, foreignCol string
		if err = rows.Scan(&col, &foreignTable, &foreignCol); err != nil {
			return nil, ctx("IntrospectTableForeignKeys", err)
		}
		constraints = append(constraints, schema.ForeignKeyConstraint{
			SourceTableField: col,
			TargetTable:      foreignTable,
			TargetTableField: foreignCol,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return constraints, nil
}

func (p PgDB) IntrospectTableUniqueConstraints(table string, cols []*schema.Column) error {
	rows, err := p.Query(`select kcu.column_name
                            from information_schema.table_constraints as tc
                            join information_schema.key_column_usage as kcu
                              on tc.constraint_name = kcu.constraint_name
                           where tc.constraint_type = 'UNIQUE'
                             and tc.table_name = $1`, table)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var col string
		if err = rows.Scan(&col); err != nil {
			return ctx("IntrospectTableUniqueConstraints", err)
		}
		if col != "id" {
			if err = AddUniqueSpecifier(cols, col); err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (p PgDB) IntrospectTableIndexes(table string, cols []*schema.Column) ([]*schema.Index, error) {
	indexNameRows, err :=
		p.Query(`select c.relname as index_name,
                        i.indkey as column_indexes
                   from pg_catalog.pg_class c
                   join pg_catalog.pg_index i on c.oid = i.indexrelid
                   join pg_catalog.pg_class c2 on i.indrelid = c2.oid
                   join pg_catalog.pg_namespace ns
                     on c.relnamespace = ns.oid
                  where c.relkind = 'i' and ns.nspname = 'public'
                    and not i.indisprimary and not i.indisunique
                    and c2.relname = $1`, table)
	if err != nil {
		return nil, err
	}
	defer indexNameRows.Close()

	indexes := []*schema.Index{}
	for indexNameRows.Next() {
		var index, colArray string

		if err = indexNameRows.Scan(&index, &colArray); err != nil {
			return nil, ctx("IntrospectTableIndexes", err)
		}

		colIndices, err := SplitIntArray(colArray)
		if err != nil {
			return nil, err
		}
		indexes =
			append(indexes, p.TableIndex(table, index, cols, colIndices))
	}
	if err = indexNameRows.Err(); err != nil {
		return nil, err
	}
	return indexes, nil
}

func (p PgDB) TableIndex(
	table, index string, cols []*schema.Column,
	indexColNumbers []int) *schema.Index {
	indexDef := schema.Index{
		Name:      index,
		TableName: table,
		Columns:   make([]string, len(indexColNumbers)),
	}
	for i, num := range indexColNumbers {
		if num == 0 {
			indexDef.Columns[i] = schema.UnknownColumn
		} else {
			indexDef.Columns[i] = cols[num-1].Name
		}
	}
	return &indexDef
}

func SplitIntArray(intarray string) ([]int, error) {
	intarray = StripArrayBraces(intarray)
	intstrings := strings.Split(intarray, " ")
	res := make([]int, len(intstrings))
	var err error
	for i, intstr := range intstrings {
		num, err := strconv.ParseInt(intstr, 10, 32)
		if err != nil {
			return nil, err
		}
		res[i] = int(num)
	}
	return res, err
}

func StripArrayBraces(bracedarray string) string {
	if bracedarray[0] == '{' {
		bracedarray = bracedarray[1:]
	}
	if bracedarray[len(bracedarray)-1] == '}' {
		bracedarray = bracedarray[:len(bracedarray)-1]
	}
	return bracedarray
}

func AddUniqueSpecifier(cols []*schema.Column, colname string) error {
	for _, col := range cols {
		if col.Name == colname {
			if strings.Index(col.SqlType, " unique") == -1 {
				col.SqlType += " unique"
			}
			return nil
		}
	}
	return fmt.Errorf("could not flag `%s` as unique: no such column in %#v",
		colname, cols)
}
