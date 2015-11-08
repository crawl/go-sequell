package pg

import "strconv"

// Binder tracks a postgres query bind variable.
type Binder int

// NotFirst returns true if this is not the first bind variable.
func (p *Binder) NotFirst() bool {
	return *p > 1
}

// Next gets the bind variable string "$1", "$2" etc. for the current bind
// variable, and increments the binder.
func (p *Binder) Next() string {
	nv := "$" + strconv.Itoa(int(*p))
	(*p)++
	return nv
}

// NewBinder creates a bind variable generator initialized to 1.
func NewBinder() *Binder {
	var binder Binder = 1
	return &binder
}
