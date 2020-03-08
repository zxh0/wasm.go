package interpreter

import (
	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
)

var _ instance.Table = (*table)(nil)

type table struct {
	_type binary.TableType
	elems []instance.Function
}

func NewTable(tt binary.TableType) instance.Table {
	return newTable(tt)
}

func newTable(tt binary.TableType) table {
	size := tt.Limits.Min
	if max := tt.Limits.Max; max > 0 {
		size = max
	}
	return table{
		_type: tt,
		elems: make([]instance.Function, size),
	}
}

func (t table) Type() binary.TableType {
	return t._type
}

func (t table) Size() uint32 {
	return uint32(len(t.elems))
}
func (t table) Grow(n uint32) {
	// TODO
}

func (t table) GetElem(idx uint32) instance.Function {
	t.checkIdx(idx)
	elem := t.elems[idx]
	if elem == nil {
		panic("uninitialized element") // TODO
	}
	return elem
}
func (t table) SetElem(idx uint32, elem instance.Function) {
	t.checkIdx(idx)
	t.elems[idx] = elem
}

func (t table) checkIdx(idx uint32) {
	if idx >= uint32(len(t.elems)) {
		panic("undefined element")
	}
}
