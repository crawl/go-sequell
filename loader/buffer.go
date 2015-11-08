package loader

import (
	"github.com/crawl/go-sequell/xlog"
)

// An XlogBuffer accumulates a list of xlogs to be loaded
type XlogBuffer struct {
	Buffer   map[string][]xlog.Xlog
	Capacity int
	Count    int
}

// NewBuffer creates a new xlog load buffer.
func NewBuffer(size int) *XlogBuffer {
	return &XlogBuffer{
		Buffer:   map[string][]xlog.Xlog{},
		Count:    0,
		Capacity: size,
	}
}

// IsFull checks if this buffer is at its max capacity.
func (b *XlogBuffer) IsFull() bool {
	return b.Count == b.Capacity
}

// Add adds x to the xlog buffer. Adding to a full buffer causes a panic.
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

// Clear discards all buffered xlogs.
func (b *XlogBuffer) Clear() {
	b.Count = 0
	for k := range b.Buffer {
		delete(b.Buffer, k)
	}
}
