package loader

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/crawl/go-sequell/crawl/ctime"
	"github.com/crawl/go-sequell/crawl/db"
	"github.com/crawl/go-sequell/crawl/xlogtools"
	"github.com/crawl/go-sequell/pg"
	"github.com/crawl/go-sequell/sources"
	"github.com/crawl/go-sequell/xlog"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const loadBufferSize = 50000

var (
	// ErrDuplicateRow means the loader found an xlogfile row exactly identical
	// to a previously inserted row
	ErrDuplicateRow = errors.New("duplicate xlog row")
)

// A Loader loads game and milestone records into Sequell's database. Loaders
// must be configured with a set of servers, a database connection, and the
// sequell database schema.
type Loader struct {
	sources.Servers
	DB               pg.DB
	Schema           *db.CrawlSchema
	Readers          []*Reader
	gameTypePrefixes map[string]string
	RowCount         int64
	LogNorm          *xlogtools.Normalizer

	tableLookups        map[string][]*TableLookup
	tableInsertFields   map[string][]*db.Field
	tableInsertKeys     map[string][]string
	tableInsertDefaults map[string][]string
	tableCopyStatements map[string]string
	buffer              *XlogBuffer
	offsetQuery         *sql.Stmt
}

// A Reader reads records suing an XlogReader (i.e. from one xlog file), from
// the source XlogSrc, and writes those records to the Table configured.
type Reader struct {
	*xlog.Reader
	*sources.XlogSrc
	Table string
}

// New creates a new loader given a database connection, server and schema
// configs, an xlog normalizer and the set of game type mappings of Crawl
// game types to their table prefixes.
func New(db pg.DB, srv sources.Servers, sch *db.CrawlSchema, norm *xlogtools.Normalizer, gameTypePrefixes map[string]string) *Loader {
	l := &Loader{
		Servers:          srv,
		DB:               db,
		Schema:           sch,
		gameTypePrefixes: gameTypePrefixes,
		LogNorm:          norm,
	}
	l.init()
	return l
}

func (l *Loader) init() {
	if l.Readers != nil {
		return
	}
	l.buffer = NewBuffer(loadBufferSize)
	xlogs := l.Servers.XlogSources()
	l.Readers = make([]*Reader, len(xlogs))
	for i, x := range xlogs {
		l.Readers[i] = &Reader{
			Reader:  xlog.NewReader(x.Server.Name, x.TargetPath, x.TargetRelPath),
			XlogSrc: x,
			Table:   l.TableName(x),
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

	findLookup := func(field *db.Field) *TableLookup {
		lookupTable := l.Schema.FindLookupTableForField(field.Name)
		if lookup, ok := lookups[lookupTable.Name]; ok {
			return lookup
		}
		lookup := NewTableLookup(lookupTable, loadBufferSize)
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
	l.tableInsertFields = make(map[string][]*db.Field, len(l.Schema.Tables))
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
		make(map[string]string, len(l.tableInsertFields)*len(l.gameTypePrefixes))
	for table, fields := range l.tableInsertFields {
		for _, prefix := range l.gameTypePrefixes {
			table := prefix + table
			l.tableCopyStatements[table] = l.copyStatement(table, fields)
		}
	}
}

func (l *Loader) copyStatement(table string, fields []*db.Field) string {
	fieldRefNames := make([]string, len(fields))
	for i, f := range fields {
		fieldRefNames[i] = f.RefName()
	}
	return pq.CopyIn(table, fieldRefNames...)
}

// TableName returns the insertion table for the given xlog source.
func (l *Loader) TableName(x *sources.XlogSrc) string {
	return l.gameTypePrefixes[x.Game] + x.Type.BaseTable()
}

// FindReader returns the Reader object given a file path, using an exact
// match for file == Reader.TargetPath.
func (l *Loader) FindReader(file string) *Reader {
	for _, r := range l.Readers {
		if r.TargetPath == file {
			return r
		}
	}
	return nil
}

// LoadLog loads outstanding logs in the given file.
func (l *Loader) LoadLog(file string) error {
	reader := l.FindReader(file)
	if reader == nil {
		return fmt.Errorf("No reader known for %s", file)
	}
	return l.LoadReaderLogs(reader)
}

// LoadCommitLog loads a single log and commits all records to the db.
func (l *Loader) LoadCommitLog(file string) error {
	if err := l.LoadLog(file); err != nil {
		return err
	}
	return l.Commit()
}

// Load loads all outstanding logs from all readers, but does not Commit() them
// automatically. After loading logs, Load closes all file handles.
func (l *Loader) Load() error {
	l.RowCount = 0
	for _, r := range l.Readers {
		defer r.Close()
		if err := l.LoadReaderLogs(r); err != nil {
			return err
		}
	}
	return nil
}

// LoadCommit loads all outstanding logs and flushes them to the database. All
// file handles will be closed at the end of this.
func (l *Loader) LoadCommit() error {
	if err := l.Load(); err != nil {
		return errors.Wrap(err, "Loader.Load")
	}
	return l.Commit()
}

// LoadReaderLogs loads logs from a single Reader. The Reader will
// remain open at the end of this call.
func (l *Loader) LoadReaderLogs(reader *Reader) error {
	seekPos, err := l.QuerySeekOffset(reader.Filename, reader.Table)
	if err != nil {
		return errors.Wrap(err, "QuerySeekOffset")
	}
	if seekPos != -1 {
		if err = reader.SeekNext(seekPos); err != nil {
			if err == xlog.ErrNoFile {
				log.Printf("Ignoring missing file: %s\n", reader.Filename)
				return nil
			}
			var help string
			if err == io.EOF {
				help = " (did the file shrink?)"
			}
			return errors.Wrapf(err, "SeekNext:%s:%d%s", reader.Filename, seekPos, help)
		}
	}

	first := true
	offset := reader.Offset
	for {
		xlogEntry, err := reader.Next()
		if err == xlog.ErrNoFile {
			log.Printf("Ignoring missing file: %s\n", reader.Filename)
			return nil
		}
		if first && (xlogEntry != nil || err != nil) {
			log.Printf("LoadLogs: %s offset=%d", reader.Filename, offset)
			first = false
		}
		if err != nil {
			return errors.Wrap(err, "reader.Next")
		}
		if xlogEntry == nil {
			return nil
		}
		if !xlogtools.ValidXlog(xlogEntry) {
			log.Printf("LoadLogs: %s offset=%s skipping bad xlog: %#v\n",
				reader.Filename, xlogEntry[":offset"], xlogEntry)
			continue
		}
		if err = l.Add(reader, xlogEntry); err != nil {
			return err
		}
	}
}

// NormalizeLog normalizes x and adds reader metadata to it.
func (l *Loader) NormalizeLog(x xlog.Xlog, reader *Reader) error {
	x["file"] = reader.Filename
	x["table"] = reader.Table
	x["base_table"] = reader.Type.BaseTable()
	x["src"] = reader.Server.Name
	x["offset"] = x[":offset"]
	delete(x, ":offset")

	var err error
	_, err = l.LogNorm.NormalizeLog(x)
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
		return errors.Wrapf(err, "NormalizeLog(%#v)", x)
	}
	return nil
}

// Add normalizes the xlog and adds it to the buffer of xlogs to be
// saved to the database.
func (l *Loader) Add(reader *Reader, x xlog.Xlog) error {
	if err := l.NormalizeLog(x, reader); err != nil {
		return err
	}

	if l.buffer.IsFull() {
		if err := l.Commit(); err != nil {
			return err
		}
	}
	l.buffer.Add(x)
	return nil
}

// Commit saves all buffered xlogs to the database and clears the buffer.
func (l *Loader) Commit() error {
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

	tx, err := l.DB.Begin()
	if err != nil {
		return nil
	}
	fail := func(err error) error {
		tx.Rollback()
		return err
	}

	deduplicatedLogs, err := l.resolveLookupFieldIds(tx, table, lookups, logs)
	if err != nil {
		return fail(errors.Wrap(err, "resolvelookups"))
	}

	if err = l.insertTableLogs(tx, table, deduplicatedLogs); err != nil {
		return fail(errors.Wrap(err, "insertTableLogs"))
	}

	deduplicatedLogCount := len(deduplicatedLogs)
	if deduplicatedLogCount < nlogs {
		log.Printf("%s: Skipped %d/%d duplicate rows", table, nlogs-deduplicatedLogCount, nlogs)
	}

	if deduplicatedLogCount == 0 {
		tx.Rollback()
		return nil
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "loadTableLogs.Commit")
	}
	l.RowCount += int64(deduplicatedLogCount)
	log.Printf("%s: Committed %d (total: %d)\n", table, deduplicatedLogCount, l.RowCount)
	return nil
}

func (l *Loader) resolveLookupFieldIds(tx *sql.Tx, destinationTable string, lookups []*TableLookup, logs []xlog.Xlog) (deduplicatedLogs []xlog.Xlog, err error) {
	if err = l.resolveLookups(tx, lookups, logs); err != nil {
		return nil, errors.Wrap(err, "resolveLookups")
	}

	deduplicatedLogs, err = l.applyLookups(lookups, logs)
	if err != nil {
		return nil, errors.Wrap(err, "resolveLookupFieldIds")
	}

	return deduplicatedLogs, nil
}

func (l *Loader) resolveLookups(tx *sql.Tx, lookups []*TableLookup, logs []xlog.Xlog) error {
	for _, l := range lookups {
		if err := l.ResolveAll(tx, logs); err != nil {
			return err
		}
	}
	return nil
}

func removeXlogLinesAtIndexes(logs []xlog.Xlog, duplicateIndexes map[int]bool) (deduplicatedLogs []xlog.Xlog) {
	if len(duplicateIndexes) == 0 {
		return logs
	}

	deduplicatedLogs = make([]xlog.Xlog, 0, len(logs)-len(duplicateIndexes))
	for i, log := range logs {
		if !duplicateIndexes[i] {
			deduplicatedLogs = append(deduplicatedLogs, log)
		}
	}
	return deduplicatedLogs
}

func (l *Loader) applyLookups(lookups []*TableLookup, logs []xlog.Xlog) (deduplicatedLogs []xlog.Xlog, err error) {
	duplicateIndexes := map[int]bool{}

	for index, x := range logs {
		for _, lookup := range lookups {
			if err := lookup.SetIDs(x); err != nil {
				if errors.Cause(err) == ErrDuplicateRow {
					duplicateIndexes[index] = true
					continue
				}

				return nil, err
			}
		}
	}

	return removeXlogLinesAtIndexes(logs, duplicateIndexes), nil
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
		return errors.Wrap(err, "Loader.insertTableLogs.Prepare")
	}

	row := make([]interface{}, len(keys))
	fileOffsets := map[string]string{}

	for _, x := range logs {
		loadXlogRow(row, keys, defaults, x)
		if _, err := st.Exec(row...); err != nil {
			return errors.Wrapf(err, "Loader.insertTableLogs.Exec(%#v)", x)
		}
		fileOffsets[x["file"]] = x["offset"]
	}

	if _, err = st.Exec(); err != nil {
		return errors.Wrap(err, "Loader.insertTableLogs.Exec()")
	}

	if err = st.Close(); err != nil {
		return errors.Wrap(err, "Loader.insertTableLogs.Close()")
	}

	if err = l.updateFileOffsets(tx, fileOffsets); err != nil {
		return errors.Wrap(err, "Loader.updateFileOffsets")
	}

	return nil
}

func (l *Loader) updateFileOffsets(tx *sql.Tx, offsets map[string]string) error {
	noffsets := len(offsets)
	if noffsets == 0 {
		return nil
	}
	sql := l.updateFileOffsetSQL(noffsets)
	values := make([]interface{}, noffsets*2)
	i := 0
	for file, offsetText := range offsets {
		values[i] = NormalizeValue(file)
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

func (l *Loader) updateFileOffsetSQL(noffset int) string {
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

// NormalizeValue replaces underscores in value with spaces for text
// values that must be accessible through Sequell's query interface.
func NormalizeValue(value string) string {
	return strings.Replace(value, "_", " ", -1)
}

func loadXlogRow(row []interface{}, keys []string, defaults []string, x xlog.Xlog) {
	for i, key := range keys {
		value := NormalizeValue(x[key])
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
	if err := l.offsetQuery.QueryRow(NormalizeValue(file)).Scan(&offset); err != nil {
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
