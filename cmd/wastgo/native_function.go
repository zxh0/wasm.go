package main

import (
	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
)

var _ instance.Function = (*nativeFunction)(nil)

type nativeFunction struct {
	t binary.FuncType
	f instance.GoFunc
}

func (nf nativeFunction) Type() binary.FuncType {
	return nf.t
}
func (nf nativeFunction) Call(args ...interface{}) (interface{}, error) {
	return nf.f(args...)
}
