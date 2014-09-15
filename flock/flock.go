// Package flock provides a wrapper around the flock syscall.
package flock

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/greensnark/go-sequell/ectx"
)

type Lock struct {
	Path string
	File *os.File
}

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

func (l *Lock) Lock(blocking bool) error {
	if l.File == nil {
		var err error
		l.File, err = os.OpenFile(l.Path, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_TRUNC, os.ModePerm)
		if err != nil {
			return ectx.Err(fmt.Sprintf("open %s", l.Path), err)
		}
	}

	if err := syscall.Flock(int(l.File.Fd()), l.lockMode(blocking)); err != nil {
		return ectx.Err(fmt.Sprintf("flock %s", l.Path), err)
	}
	l.File.WriteString(strconv.Itoa(os.Getpid()) + "\n")
	return nil
}

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
