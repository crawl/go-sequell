package loader

import (
	"database/sql"
	"log"
	"sync"
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

const LoadBufferSize = 5

type Loader struct {
	*sources.Servers
	DB               pg.DB
	Schema           *cdb.CrawlSchema
	Readers          []Reader
	GameTypePrefixes map[string]string

	tableLookups        map[string][]*TableLookup
	tableInsertFields   map[string][]*cdb.Field
	tableInsertKeys     map[string][]string
	tableInsertDefaults map[string][]string
	tableCopyStatements map[string]string
	buffer              *XlogBuffer
	lock                sync.Mutex
}

type Reader struct {
	*xlog.XlogReader
	*sources.XlogSrc
	Table string
}

func New(db pg.DB, srv *sources.Servers, sch *cdb.CrawlSchema, gameTypePrefixes map[string]string) *Loader {
	return &Loader{
		Servers:          srv,
		DB:               db,
		Schema:           sch,
		GameTypePrefixes: gameTypePrefixes,
	}
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
	l.lock.Lock()
	defer l.lock.Unlock()
	l.init()
	return l.Readers
}

// Load loads all outstanding logs from all readers, but does not Flush() them
// automatically
func (l *Loader) Load() error {
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
			return err
		}
	}

	for {
		xlog, err := reader.Next()
		if err != nil {
			return err
		}
		if xlog == nil {
			return nil
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

	return err
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
	if len(logs) == 0 {
		return nil
	}
	txn, err := l.DB.Begin()
	if err != nil {
		return nil
	}
	fail := func(err error) error {
		txn.Rollback()
		return err
	}
	lookups := l.tableLookups[logs[0]["base_table"]]
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
	for _, x := range logs {
		loadXlogRow(row, keys, defaults, x)
		if _, err := st.Exec(row...); err != nil {
			return ectx.Err("Loader.insertTableLogs.Exec(...)", err)
		}
	}

	if _, err = st.Exec(); err != nil {
		return ectx.Err("Loader.insertTableLogs.Exec()", err)
	}

	if err = st.Close(); err != nil {
		return ectx.Err("Loader.insertTableLogs.Close()", err)
	}

	return nil
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
	offsetQuery := "select max(file_offset) from " + table +
		" where file_id = (select id from l_file where file = $1)"
	err := l.DB.QueryRow(offsetQuery, file).Scan(&offset)
	if err != nil {
		return -1, err
	}
	if offset.Valid {
		return offset.Int64, nil
	}
	return -1, nil
}

// Close closes the loader and associated resources.
func (l *Loader) Close() error {
	l.lock.Lock()
	defer l.lock.Unlock()
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
