package interpreter

import (
	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
)

var _ instance.Global = (*globalVar)(nil)

type globalVar struct {
	_type binary.GlobalType
	val   uint64
}

func NewGlobal(gt binary.GlobalType, val uint64) instance.Global {
	return newGlobal(gt, val)
}

func newGlobal(gt binary.GlobalType, val uint64) *globalVar {
	return &globalVar{_type: gt, val: val}
}

func (g *globalVar) Type() binary.GlobalType {
	return g._type
}

func (g *globalVar) Get() uint64 {
	return g.val
}
func (g *globalVar) Set(val uint64) {
	if g._type.Mut != 1 {
		panic("constant global!")
	}
	g.val = val
}
