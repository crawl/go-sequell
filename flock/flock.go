// Package flock provides a wrapper around the flock syscall.
package flock

import (
	"os"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
)

// A Lock represents a lock file. Locks are not safe for concurrent use.
type Lock struct {
	Path string
	File *os.File
}

// New creates a file lock
func New(file string) *Lock {
	return &Lock{Path: file}
}

func (l *Lock) lockMode(blocking bool) int {
	mode := syscall.LOCK_EX
	if !blocking {
		mode = mode | syscall.LOCK_NB
	}
	return mode
}

// Lock locks the file at l.Path, blocking if necessary.
func (l *Lock) Lock(blocking bool) error {
	if l.File == nil {
		var err error
		l.File, err = os.OpenFile(l.Path, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_TRUNC, 0600)
		if err != nil {
			return errors.Wrapf(err, "open %s", l.Path)
		}
	}

	if err := syscall.Flock(int(l.File.Fd()), l.lockMode(blocking)); err != nil {
		return errors.Wrapf(err, "flock %s", l.Path)
	}
	l.File.WriteString(strconv.Itoa(os.Getpid()) + "\n")
	return nil
}

// Unlock unlocks a locked file.
func (l *Lock) Unlock() error {
	if l.File == nil {
		return nil
	}
	defer func() {
		l.File.Close()
		os.Remove(l.Path)
		l.File = nil
	}()
	return syscall.Flock(int(l.File.Fd()), syscall.LOCK_UN)
}
