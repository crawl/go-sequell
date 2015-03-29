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

var ErrNoFile = errors.New("xlog file not found")

type XlogReader struct {
	Path     string
	Filename string
	File     *os.File
	Offset   int64
	Reader   *bufio.Reader
}

// Reader creates a new Xlog reader for the given absolute path, and
// the given dbFilename. The dbFilename will be saved as the filename
// in the database.
func Reader(filepath, dbFilename string) *XlogReader {
	return &XlogReader{Path: filepath, Filename: dbFilename}
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

func (x *XlogReader) String() string {
	var opened string
	if x.File != nil {
		opened = " opened"
	}
	return fmt.Sprintf("XlogReader[path=%s%s offset=%d]",
		x.Path, opened, x.Offset)
}

func (x *XlogReader) Close() error {
	x.Reader = nil
	if x.File != nil {
		if err := x.File.Close(); err != nil {
			return err
		}
		x.File = nil
	}
	return nil
}

func (x *XlogReader) open() error {
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

// Seek seeks to the given offset from the start of the file.
func (x *XlogReader) Seek(offset int64) error {
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
func (x *XlogReader) SeekNext(offset int64) error {
	if err := x.Seek(offset); err != nil {
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
func (x *XlogReader) BackToLastCompleteLine() error {
	if x.File != nil {
		return x.Seek(x.Offset)
	}
	return nil
}

// ReadAll reads all available Xlog lines from the source; use only for testing.
func (x *XlogReader) ReadAll() ([]Xlog, error) {
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
func (x *XlogReader) Next() (Xlog, error) {
	if err := x.open(); err != nil {
		return nil, err
	}
	var readOffset int64 = 0
	for {
		line, err := x.ReadCompleteLine()
		if err != nil && err != io.EOF {
			return nil, err
		}
		var lineLen int64 = int64(len(line))
		readOffset += lineLen

		// Discard trailing newline and whitespace
		if lineLen > 0 {
			line = strings.TrimRight(line, " \n\r\t")
		}
		if !IsPotentialXlogLine(line) {
			if err == nil {
				continue
			}
			// EOF with no parseable data: go back to last line.
			err = x.BackToLastCompleteLine()
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		parsedXlog, err := Parse(line)
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
func (x *XlogReader) ReadCompleteLine() (string, error) {
	line, err := x.Reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return line, nil
}
