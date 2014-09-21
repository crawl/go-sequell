package pg

import "strconv"

type PGBinder int

func (p *PGBinder) NotFirst() bool {
	return *p > 1
}

func (p *PGBinder) Next() string {
	nv := "$" + strconv.Itoa(int(*p))
	(*p)++
	return nv
}

func NewBinder() *PGBinder {
	var binder PGBinder = 1
	return &binder
}
