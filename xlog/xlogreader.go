package xlog

import (
	"bufio"
	"errors"
	"io"
	"os"
	"syscall"
)

var ErrNoFile = errors.New("xlog file not found")

type XlogReader struct {
	Path   string
	File   *os.File
	Offset int64
	Reader *bufio.Reader
}

func Reader(path string) *XlogReader {
	return &XlogReader{Path: path}
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

func (x *XlogReader) Close() error {
	if x.File != nil {
		if err := x.File.Close(); err != nil {
			return err
		}
		x.File = nil
		x.Reader = nil
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
		return translateErr(err)
	}
	x.Reader = bufio.NewReader(x.File)
	return nil
}

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

// BackToLastCompleteLine rewinds the XlogReader to the end of the
// last complete line read, or the last place explicitly Seek()ed to;
// does nothing if nothing read yet.
func (x *XlogReader) BackToLastCompleteLine() error {
	if x.File != nil {
		return x.Seek(x.Offset)
	}
	return nil
}

// ReadAll reads all available Xlog lines from the source. Can potentially
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
		readOffset += int64(len(line))
		if !IsPotentialXlogLine(line) {
			if err == nil {
				continue
			}
			// EOF, no parseable data:
			err = x.BackToLastCompleteLine()
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		parsedXlog, err := Parse(line)
		if err != nil {
			x.BackToLastCompleteLine()
			return nil, err
		}
		x.Offset += readOffset
		return parsedXlog, nil
	}
}

// ReadCompleteLine reads a complete line from the Xlog reader,
// returning an empty string if it can't read a full \n-terminated
// line.
func (x *XlogReader) ReadCompleteLine() (string, error) {
	line, err := x.Reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return line, err
}
