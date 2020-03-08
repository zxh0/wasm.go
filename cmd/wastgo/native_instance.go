package main

import (
	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
	"github.com/zxh0/wasm.go/interpreter"
)

var _ instance.Instance = (*nativeInstance)(nil)

type nativeInstance struct {
	exported map[string]interface{}
}

func newNativeInstance() *nativeInstance {
	return &nativeInstance{
		exported: map[string]interface{}{},
	}
}

func (n *nativeInstance) RegisterFunc(name string,
	ft binary.FuncType, f instance.GoFunc) {

	n.exported[name] = nativeFunction{t: ft, f: f}
}
func (n *nativeInstance) RegisterNoResultsFunc(name string,
	f instance.GoFunc, params ...binary.ValType) {

	ft := binary.FuncType{
		ParamTypes:  params,
		ResultTypes: []binary.ValType{},
	}
	n.RegisterFunc(name, ft, f)
}
func (n *nativeInstance) RegisterGlobal(name string,
	vt binary.ValType, mut bool, val uint64) {
	gt := binary.GlobalType{ValType: vt}
	if mut {
		gt.Mut = 1
	}
	n.exported[name] = interpreter.NewGlobal(gt, val)
}
func (n *nativeInstance) RegisterTable(name string, min, max uint32) {
	tt := binary.TableType{
		ElemType: binary.FuncRef,
		Limits:   binary.Limits{Min: min, Max: max},
	}
	n.exported[name] = interpreter.NewTable(tt)
}
func (n *nativeInstance) RegisterMem(name string, min, max uint32) {
	mt := binary.MemType{Min: min, Max: max}
	n.exported[name] = interpreter.NewMemory(mt)
}

func (n *nativeInstance) GetGlobalValue(name string) (interface{}, error) {
	panic("implement me")
}

func (n *nativeInstance) CallFunc(name string, args ...interface{}) (interface{}, error) {
	panic("implement me")
}

func (n *nativeInstance) Get(name string) interface{} {
	return n.exported[name]
}
