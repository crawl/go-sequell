package loader

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/greensnark/go-sequell/crawl/ctime"
	cdb "github.com/greensnark/go-sequell/crawl/db"
	"github.com/greensnark/go-sequell/crawl/xlogtools"
	"github.com/greensnark/go-sequell/ectx"
	"github.com/greensnark/go-sequell/pg"
	"github.com/greensnark/go-sequell/sources"
	"github.com/greensnark/go-sequell/xlog"
	"github.com/lib/pq"
)

const LoadBufferSize = 50000

type Loader struct {
	*sources.Servers
	DB               pg.DB
	Schema           *cdb.CrawlSchema
	Readers          []Reader
	GameTypePrefixes map[string]string
	RowCount         int64

	tableLookups        map[string][]*TableLookup
	tableInsertFields   map[string][]*cdb.Field
	tableInsertKeys     map[string][]string
	tableInsertDefaults map[string][]string
	tableCopyStatements map[string]string
	buffer              *XlogBuffer
	offsetQuery         *sql.Stmt
}

type Reader struct {
	*xlog.XlogReader
	*sources.XlogSrc
	Table string
}

func New(db pg.DB, srv *sources.Servers, sch *cdb.CrawlSchema, gameTypePrefixes map[string]string) *Loader {
	l := &Loader{
		Servers:          srv,
		DB:               db,
		Schema:           sch,
		GameTypePrefixes: gameTypePrefixes,
	}
	l.init()
	return l
}

func (l *Loader) init() {
	if l.Readers != nil {
		return
	}
	l.buffer = NewBuffer(LoadBufferSize)
	xlogs := l.Servers.XlogSources()
	l.Readers = make([]Reader, len(xlogs))
	for i, x := range xlogs {
		l.Readers[i] = Reader{
			XlogReader: xlog.Reader(x.TargetPath),
			XlogSrc:    x,
			Table:      l.TableName(x),
		}
	}
	l.createTableLookups()
	l.initTableInsertFields()
	l.initCopyStatements()
	l.offsetQuery = l.createOffsetQueryStmt()
}

func (l *Loader) createOffsetQueryStmt() *sql.Stmt {
	stmt, err := l.DB.Prepare("select file_offset from l_file where file = $1")
	if err != nil {
		panic(err)
	}
	return stmt
}

func (l *Loader) createTableLookups() {
	lookups := map[string]*TableLookup{}

	findLookup := func(field *cdb.Field) *TableLookup {
		lookupTable := l.Schema.FindLookupTableForField(field.Name)
		if lookup, ok := lookups[lookupTable.Name]; ok {
			return lookup
		}
		lookup := NewTableLookup(lookupTable, LoadBufferSize)
		lookups[lookupTable.Name] = lookup
		return lookup
	}

	l.tableLookups = map[string][]*TableLookup{}
	for _, baseTable := range l.Schema.Tables {
		tableName := baseTable.Name
		foundLookups := map[string]bool{}
		for _, f := range baseTable.Fields {
			if f.ForeignKeyLookup {
				tableLookup := findLookup(f)
				if !foundLookups[tableLookup.Name()] {
					foundLookups[tableLookup.Name()] = true
					l.tableLookups[tableName] =
						append(l.tableLookups[tableName], tableLookup)
				}
			}
		}
	}
}

func (l *Loader) initTableInsertFields() {
	l.tableInsertFields = make(map[string][]*cdb.Field, len(l.Schema.Tables))
	l.tableInsertKeys = make(map[string][]string, len(l.Schema.Tables))
	l.tableInsertDefaults = make(map[string][]string, len(l.Schema.Tables))
	for _, baseTable := range l.Schema.Tables {
		// Skip id field:
		fields := baseTable.Fields[1:]
		l.tableInsertFields[baseTable.Name] = fields

		keys := make([]string, len(fields))
		defaults := make([]string, len(fields))
		for i, f := range fields {
			keys[i] = f.ResolvedKey()
			defaults[i] = f.DefaultValue
		}
		l.tableInsertKeys[baseTable.Name] = keys
		l.tableInsertDefaults[baseTable.Name] = defaults
	}
}

func (l *Loader) initCopyStatements() {
	l.tableCopyStatements =
		make(map[string]string, len(l.tableInsertFields)*len(l.GameTypePrefixes))
	for table, fields := range l.tableInsertFields {
		for _, prefix := range l.GameTypePrefixes {
			table := prefix + table
			l.tableCopyStatements[table] = l.copyStatement(table, fields)
		}
	}
}

func (l *Loader) copyStatement(table string, fields []*cdb.Field) string {
	fieldRefNames := make([]string, len(fields))
	for i, f := range fields {
		fieldRefNames[i] = f.RefName()
	}
	return pq.CopyIn(table, fieldRefNames...)
}

// TableName returns the insertion table for the given source.
func (l *Loader) TableName(x *sources.XlogSrc) string {
	return l.GameTypePrefixes[x.Game] + x.Type.BaseTable()
}

func (l *Loader) getReaders() []Reader {
	return l.Readers
}

// Load loads all outstanding logs from all readers, but does not Flush() them
// automatically
func (l *Loader) Load() error {
	l.RowCount = 0
	readers := l.getReaders()
	for _, r := range readers {
		if err := l.LoadReaderLogs(r); err != nil {
			return err
		}
	}
	return nil
}

// LoadCommit loads all outstanding logs and flushes them to the database.
func (l *Loader) LoadCommit() error {
	if err := l.Load(); err != nil {
		return ectx.Err("Loader.Load", err)
	}
	return l.Flush()
}

func (l *Loader) LoadReaderLogs(reader Reader) error {
	seekPos, err := l.QuerySeekOffset(reader.Filename, reader.Table)
	if err != nil {
		return ectx.Err("QuerySeekOffset", err)
	}
	if seekPos != -1 {
		if err = reader.SeekNext(seekPos); err != nil {
			if err == xlog.ErrNoFile {
				log.Printf("Ignoring missing file: %s\n", reader)
				return nil
			}
			return ectx.Err("SeekNext", err)
		}
	}

	first := true
	offset := reader.Offset
	for {
		xlog, err := reader.Next()
		if first && (xlog != nil || err != nil) {
			log.Printf("LoadLogs: %s offset=%d", reader.Filename, offset)
			first = false
		}
		if err != nil {
			return ectx.Err("reader.Next", err)
		}
		if xlog == nil {
			return nil
		}
		if !xlogtools.ValidXlog(xlog) {
			log.Printf("LoadLogs: %s offset=%s skipping bad xlog: %#v\n",
				reader.Filename, xlog[":offset"], xlog)
			continue
		}
		if err = l.Add(reader, xlog); err != nil {
			return err
		}
	}
}

func (l *Loader) NormalizeLog(x xlog.Xlog, reader Reader) error {
	x["file"] = reader.Filename
	x["table"] = reader.Table
	x["base_table"] = reader.Type.BaseTable()
	x["src"] = reader.Server.Name
	x["offset"] = x[":offset"]
	delete(x, ":offset")

	var err error
	_, err = xlogtools.NormalizeLog(x)
	if err != nil {
		return err
	}

	normTime := func(field string) {
		if timeStr, ok := x[field]; ok {
			var t time.Time
			if t, err = reader.Server.ParseLogTime(timeStr); err == nil {
				x[field] = t.Format(ctime.LayoutDBTime)
			}
		}
	}
	normTime("start")
	normTime("end")
	normTime("time")

	if err != nil {
		return ectx.Err(fmt.Sprintf("NormalizeLog(%#v)", x), err)
	}
	return nil
}

// Add normalizes the xlog and adds it to the buffer of xlogs to be
// saved to the database.
func (l *Loader) Add(reader Reader, x xlog.Xlog) error {
	if err := l.NormalizeLog(x, reader); err != nil {
		return err
	}

	if l.buffer.IsFull() {
		if err := l.Flush(); err != nil {
			return err
		}
	}
	l.buffer.Add(x)
	return nil
}

// Flush saves all buffered xlogs to the database and clears the buffer.
func (l *Loader) Flush() error {
	if err := l.saveBufferedLogs(); err != nil {
		return err
	}
	l.buffer.Clear()
	return nil
}

func (l *Loader) saveBufferedLogs() error {
	for table, xlogs := range l.buffer.Buffer {
		if err := l.loadTableLogs(table, xlogs); err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) loadTableLogs(table string, logs []xlog.Xlog) error {
	nlogs := len(logs)
	if nlogs == 0 {
		return nil
	}

	lookups := l.tableLookups[logs[0]["base_table"]]

	txn, err := l.DB.Begin()
	if err != nil {
		return nil
	}
	fail := func(err error) error {
		txn.Rollback()
		return err
	}
	l.queueLookups(lookups, logs)
	if err = l.resolveLookups(txn, lookups); err != nil {
		return fail(ectx.Err("resolveLookups", err))
	}
	if err := l.applyLookups(lookups, logs); err != nil {
		return fail(ectx.Err("applyLookups", err))
	}
	if err := l.insertTableLogs(txn, table, logs); err != nil {
		return fail(ectx.Err("insertTableLogs", err))
	}
	if err := txn.Commit(); err != nil {
		return ectx.Err("loadTableLogs.Commit", err)
	}
	l.RowCount += int64(nlogs)
	log.Printf("%s: Committed %d (total: %d)\n", table, nlogs, l.RowCount)
	return nil
}

func (l *Loader) queueLookups(lookups []*TableLookup, logs []xlog.Xlog) {
	for _, x := range logs {
		for _, lookup := range lookups {
			lookup.Add(x)
		}
	}
}

func (l *Loader) resolveLookups(tx *sql.Tx, lookups []*TableLookup) error {
	for _, l := range lookups {
		if err := l.ResolveAll(tx); err != nil {
			return err
		}
	}
	return nil
}

func (l *Loader) applyLookups(lookups []*TableLookup, logs []xlog.Xlog) error {
	for _, x := range logs {
		for _, lookup := range lookups {
			if err := lookup.SetIds(x); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Loader) insertTableLogs(tx *sql.Tx, table string, logs []xlog.Xlog) error {
	if len(logs) == 0 {
		return nil
	}

	baseTable := logs[0]["base_table"]
	keys := l.tableInsertKeys[baseTable]
	defaults := l.tableInsertDefaults[baseTable]
	st, err := tx.Prepare(l.tableCopyStatements[table])
	if err != nil {
		return ectx.Err("Loader.insertTableLogs.Prepare", err)
	}

	row := make([]interface{}, len(keys))
	fileOffsets := map[string]string{}

	for _, x := range logs {
		loadXlogRow(row, keys, defaults, x)
		if _, err := st.Exec(row...); err != nil {
			return ectx.Err(
				fmt.Sprintf("Loader.insertTableLogs.Exec(%#v)", x), err)
		}
		fileOffsets[x["file"]] = x["offset"]
	}

	if _, err = st.Exec(); err != nil {
		return ectx.Err("Loader.insertTableLogs.Exec()", err)
	}

	if err = st.Close(); err != nil {
		return ectx.Err("Loader.insertTableLogs.Close()", err)
	}

	if err = l.updateFileOffsets(tx, fileOffsets); err != nil {
		return ectx.Err("Loader.updateFileOffsets", err)
	}

	return nil
}

func (l *Loader) updateFileOffsets(tx *sql.Tx, offsets map[string]string) error {
	noffsets := len(offsets)
	if noffsets == 0 {
		return nil
	}
	sql := l.updateFileOffsetSql(noffsets)
	values := make([]interface{}, noffsets*2)
	i := 0
	for file, offsetText := range offsets {
		values[i] = file
		offset, err := strconv.ParseInt(offsetText, 10, 64)
		if err != nil {
			return err
		}
		values[i+1] = offset
		i += 2
	}
	_, err := tx.Exec(sql, values...)
	return err
}

func (l *Loader) updateFileOffsetSql(noffset int) string {
	var buf bytes.Buffer
	buf.WriteString(`update l_file f set file_offset = c.file_offset
                              from (values `)
	index := 0
	for i := 0; i < noffset; i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString("($")
		buf.WriteString(strconv.Itoa(index + 1))
		index++
		buf.WriteString(", $")
		buf.WriteString(strconv.Itoa(index + 1))
		buf.WriteString("::bigint")
		index++
		buf.WriteString(")")
	}
	buf.WriteString(`) as c (file, file_offset) where f.file = c.file`)
	return buf.String()
}

func loadXlogRow(row []interface{}, keys []string, defaults []string, x xlog.Xlog) {
	for i, key := range keys {
		value := x[key]
		if value == "" {
			value = defaults[i]
		}
		row[i] = value
	}
}

// QuerySeekOffset checks the last read offset of the file as saved in
// the table, or -1 if the file is not referenced in the table.
func (l *Loader) QuerySeekOffset(file, table string) (int64, error) {
	var offset sql.NullInt64
	if err := l.offsetQuery.QueryRow(file).Scan(&offset); err != nil {
		if err == sql.ErrNoRows {
			return -1, nil
		}
		return -1, err
	}
	if offset.Valid {
		return offset.Int64, nil
	}
	return -1, nil
}

// Close closes the loader and associated resources.
func (l *Loader) Close() error {
	if l.Readers == nil {
		return nil
	}
	for _, r := range l.Readers {
		r.Close()
	}
	return nil
}

// Monitor monitors all known logs for changes and incrementally loads
// those when they change.
func (l *Loader) Monitor() error {
	return nil
}
