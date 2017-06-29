package xlog

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// ErrNoFile means an attempt was made to read a missing xlog
var ErrNoFile = errors.New("xlog file not found")

// A Reader reads xlog entries from a logfile.
type Reader struct {
	// SourceKey is a unique identifier for the server this logfile is from
	SourceKey string

	Path     string
	Filename string
	File     *os.File
	Offset   int64
	Reader   *bufio.Reader
}

// NewReader creates a new Reader for the given absolute path, and dbFilename.
// The dbFilename will be saved as the filename in the database.
func NewReader(sourceKey, filepath, dbFilename string) *Reader {
	return &Reader{
		SourceKey: sourceKey,
		Path:      filepath,
		Filename:  dbFilename,
	}
}

func translateErr(e error) error {
	if err, ok := e.(*os.PathError); ok {
		errCode := err.Err
		if errCode == syscall.ENOENT {
			return ErrNoFile
		}
	}
	return e
}

func (x *Reader) String() string {
	var opened string
	if x.File != nil {
		opened = " opened"
	}
	return fmt.Sprintf("XlogReader[path=%s%s offset=%d]",
		x.Path, opened, x.Offset)
}

// Close closes the Reader's file handle.
func (x *Reader) Close() error {
	x.Reader = nil
	if x.File != nil {
		if err := x.File.Close(); err != nil {
			return err
		}
		x.File = nil
	}
	return nil
}

func (x *Reader) open() error {
	if x.File != nil {
		return nil
	}
	var err error
	x.File, err = os.Open(x.Path)
	if err != nil {
		log.Println("Error opening file:", x.Path, err)
		return translateErr(err)
	}
	x.Reader = bufio.NewReader(x.File)
	return nil
}

// SeekOffset seeks to the given offset from the start of the file.
func (x *Reader) SeekOffset(offset int64) error {
	if err := x.open(); err != nil {
		return err
	}
	_, err := x.File.Seek(offset, 0)
	if err == nil {
		x.Offset = offset
		x.Reader.Reset(x.File)
	}
	return err
}

// SeekNext seeks to the given offset, then reads and discards one
// complete line. This is convenient if you have the offset of the
// last processed line and want to resume reading on the next line.
func (x *Reader) SeekNext(offset int64) error {
	if err := x.SeekOffset(offset); err != nil {
		return err
	}
	line, err := x.ReadCompleteLine()
	if err == nil {
		x.Offset += int64(len(line))
		return nil
	}
	return err
}

// BackToLastCompleteLine rewinds the XlogReader to the end of the
// last complete line read, or the last place explicitly Seek()ed to;
// does nothing if nothing read yet.
func (x *Reader) BackToLastCompleteLine() error {
	if x.File == nil {
		return nil
	}
	return x.SeekOffset(x.Offset)
}

// ReadAll reads all available Xlog lines from the source; use only for testing.
func (x *Reader) ReadAll() ([]Xlog, error) {
	res := []Xlog{}
	for {
		xlog, err := x.Next()
		if err != nil {
			return nil, err
		}
		if xlog == nil {
			break
		}
		res = append(res, xlog)
	}
	return res, nil
}

// Next reads the next xlog entry from the logfile, skipping blank
// lines. When EOF is reached, returns nil with no error.
func (x *Reader) Next() (Xlog, error) {
	if err := x.open(); err != nil {
		return nil, err
	}
	var readOffset int64
	for {
		line, err := x.ReadCompleteLine()
		if err != nil && err != io.EOF {
			return nil, err
		}

		atEOF := err == io.EOF
		lineLen := int64(len(line))
		readOffset += lineLen

		// Discard trailing newline and whitespace
		if lineLen > 0 {
			line = strings.TrimRight(line, " \n\r\t")
		}
		if !IsPotentialXlogLine(line) {
			if !atEOF {
				continue
			}

			// EOF with no parseable data: go back to last line.
			return nil, x.BackToLastCompleteLine()
		}

		parsedXlog, err := Parse(line, x.SourceKey)
		if err != nil {
			log.Printf("Xlog %s:%d skipping malformed line %#v\n",
				x.Path, x.Offset+readOffset-lineLen, line)
			continue
		}
		x.Offset += readOffset
		parsedXlog[":offset"] = strconv.FormatInt(x.Offset-lineLen, 10)
		return parsedXlog, nil
	}
}

// ReadCompleteLine reads a complete line from the Xlog reader,
// returning an empty string if it can't read a full \n-terminated
// line. The line returned is always empty or \n-terminated.
func (x *Reader) ReadCompleteLine() (string, error) {
	line, err := x.Reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return line, nil
}
