package pg

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/crawl/go-sequell/schema"
	"github.com/pkg/errors"
)

// A PID is the PID of a postgres client connection; this is not the same as
// a Unix process ID.
type PID int

// ActiveConnections gets the list of client PIDs connected to the DB.
func (p DB) ActiveConnections(db string) ([]PID, error) {
	rows, err :=
		p.Query(`select pid from pg_stat_activity
				  where pid != pg_backend_pid()
					and datname = $1`, db)
	if err != nil {
		return nil, errors.Wrap(err, "pg_stat_activity")
	}
	defer rows.Close()

	pids := []PID{}
	for rows.Next() {
		var pid int
		if err := rows.Scan(&pid); err != nil {
			return nil, errors.Wrap(err, "pg_stat_activity.scan")
		}
		pids = append(pids, PID(pid))
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return pids, nil
}

// TerminateConnection kills the connection specified by pid.
func (p DB) TerminateConnection(pid PID) error {
	_, err := p.Exec(`select pg_terminate_backend($1)`, int(pid))
	return errors.Wrap(err, "pg_terminate_backend")
}

// IntrospectSchema discovers the database schema.
func (p DB) IntrospectSchema() (*schema.Schema, error) {
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

// IntrospectTableNames discovers the names of tables in the db.
func (p DB) IntrospectTableNames() ([]string, error) {
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

// IntrospectTableSchemas discovers the structure of the named tables in the db.
func (p DB) IntrospectTableSchemas(tables []string) (schemas []*schema.Table, err error) {
	schemas = make([]*schema.Table, len(tables))
	for i, t := range tables {
		schemas[i], err = p.IntrospectTable(t)
		if err != nil {
			return nil, err
		}
	}
	return schemas, nil
}

// IntrospectTable discovers the structure of table.
func (p DB) IntrospectTable(table string) (*schema.Table, error) {
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

// IntrospectTableColumns introspects the columns in table from the db.
func (p DB) IntrospectTableColumns(table string) ([]*schema.Column, error) {
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
		cols = append(cols, p.columnDef(colName.String, colDefault.String, dataType.String, numericPrecision.Int64))
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return cols, nil
}

func (p DB) columnDef(name, defval, dataType string, precision int64) *schema.Column {
	dataType, defval = p.normalizeType(dataType, defval, precision)
	if defval != "" {
		defval = "default " + defval
	}
	return &schema.Column{
		Name:    name,
		SQLType: dataType,
		Default: defval,
	}
}

// normalizeType normalizes an introspected type.
func (p DB) normalizeType(sqlType, defval string, precision int64) (string, string) {
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

// IntrospectTableConstraints discovers the constraints on table with cols.
func (p DB) IntrospectTableConstraints(table string, cols []*schema.Column) ([]schema.Constraint, error) {
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

// IntrospectTablePrimaryKey discovers the primary key for a table.
func (p DB) IntrospectTablePrimaryKey(table string) (schema.Constraint, error) {
	row := p.QueryRow(`select kcu.column_name, tc.constraint_name
							  from information_schema.table_constraints as tc
							  join information_schema.key_column_usage as kcu
								on tc.constraint_name = kcu.constraint_name
							 where tc.constraint_type = 'PRIMARY KEY'
							   and tc.table_name = $1`, table)
	var pkey string
	var constraintName sql.NullString
	if err := row.Scan(&pkey, &constraintName); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, ctx("IntrospectTablePrimaryKey", err)
	}
	return schema.PrimaryKeyConstraint{
		ConstraintName: constraintName.String,
		Column:         pkey,
	}, nil
}

// IntrospectTableForeignKeys discovers the foreign keys for table.
func (p DB) IntrospectTableForeignKeys(table string) ([]schema.Constraint, error) {
	rows, err := p.Query(`select kcu.column_name,
								 ccu.table_name as foreign_table,
								 ccu.column_name as foreign_column,
								 tc.constraint_name
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
		var constraintName sql.NullString
		if err = rows.Scan(&col, &foreignTable, &foreignCol, &constraintName); err != nil {
			return nil, ctx("IntrospectTableForeignKeys", err)
		}
		constraints = append(constraints, schema.ForeignKeyConstraint{
			ConstraintName:   constraintName.String,
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

// IntrospectTableUniqueConstraints discovers the unique constraints for table
// with cols.
func (p DB) IntrospectTableUniqueConstraints(table string, cols []*schema.Column) error {
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

// IntrospectTableIndexes discovers the indexes for table with cols.
func (p DB) IntrospectTableIndexes(table string, cols []*schema.Column) ([]*schema.Index, error) {
	indexNameRows, err :=
		p.Query(`select c.relname as index_name,
						i.indkey as column_indexes,
						i.indisunique
				   from pg_catalog.pg_class c
				   join pg_catalog.pg_index i on c.oid = i.indexrelid
				   join pg_catalog.pg_class c2 on i.indrelid = c2.oid
				   join pg_catalog.pg_namespace ns
					 on c.relnamespace = ns.oid
				  where c.relkind = 'i' and ns.nspname = 'public'
					and not i.indisprimary
					and c2.relname = $1`, table)
	if err != nil {
		return nil, err
	}
	defer indexNameRows.Close()

	indexes := []*schema.Index{}
	for indexNameRows.Next() {
		var index, colArray string
		var unique bool

		if err = indexNameRows.Scan(&index, &colArray, &unique); err != nil {
			return nil, ctx("IntrospectTableIndexes", err)
		}

		colIndices, err := splitIntArray(colArray)
		if err != nil {
			return nil, err
		}
		indexes =
			append(indexes, p.tableIndex(table, index, unique, cols, colIndices))
	}
	if err = indexNameRows.Err(); err != nil {
		return nil, err
	}
	return indexes, nil
}

func (p DB) tableIndex(
	table, index string, unique bool, cols []*schema.Column,
	indexColNumbers []int) *schema.Index {
	indexDef := schema.Index{
		Name:      index,
		TableName: table,
		Columns:   make([]string, len(indexColNumbers)),
		Unique:    unique,
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

func splitIntArray(intarray string) ([]int, error) {
	intarray = stripArrayBraces(intarray)
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

func stripArrayBraces(bracedarray string) string {
	if bracedarray[0] == '{' {
		bracedarray = bracedarray[1:]
	}
	if bracedarray[len(bracedarray)-1] == '}' {
		bracedarray = bracedarray[:len(bracedarray)-1]
	}
	return bracedarray
}

// AddUniqueSpecifier adds unique to the type of a column referenced by a
// unique index.
func AddUniqueSpecifier(cols []*schema.Column, colname string) error {
	for _, col := range cols {
		if col.Name == colname {
			if strings.Index(col.SQLType, " unique") == -1 {
				col.SQLType += " unique"
			}
			return nil
		}
	}
	return fmt.Errorf("could not flag `%s` as unique: no such column in %#v",
		colname, cols)
}
