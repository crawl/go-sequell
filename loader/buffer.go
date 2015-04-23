package loader

import (
	"github.com/crawl/go-sequell/xlog"
)

type XlogBuffer struct {
	Buffer   map[string][]xlog.Xlog
	Capacity int
	Count    int
}

func NewBuffer(size int) *XlogBuffer {
	return &XlogBuffer{
		Buffer:   map[string][]xlog.Xlog{},
		Count:    0,
		Capacity: size,
	}
}

func (b *XlogBuffer) IsFull() bool {
	return b.Count == b.Capacity
}

func (b *XlogBuffer) Add(x xlog.Xlog) {
	if b.IsFull() {
		panic("buffer overflow")
	}
	table := x["table"]
	var slice []xlog.Xlog
	if slice = b.Buffer[table]; slice == nil {
		slice = make([]xlog.Xlog, 0, b.Capacity)
	}
	b.Buffer[table] = append(slice, x)
	b.Count++
}

func (b *XlogBuffer) Clear() {
	b.Count = 0
	for k := range b.Buffer {
		delete(b.Buffer, k)
	}
}
