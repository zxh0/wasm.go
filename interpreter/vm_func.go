package interpreter

import (
	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
)

var _ instance.Function = (*vmFunc)(nil)

type vmFunc struct {
	vm       *vm
	_type    binary.FuncType
	code     binary.Code
	imported instance.Function
}

func newExternalFunc(vm *vm, ft binary.FuncType,
	f instance.Function) vmFunc {

	return vmFunc{
		vm:       vm,
		_type:    ft,
		imported: f,
	}
}
func newInternalFunc(vm *vm, ft binary.FuncType,
	code binary.Code) vmFunc {

	return vmFunc{
		vm:    vm,
		_type: ft,
		code:  code,
	}
}

func (f vmFunc) Type() binary.FuncType {
	return f._type
}
func (f vmFunc) Call(args ...interface{}) (interface{}, error) {
	if f.imported != nil {
		return f.imported.Call(args...)
	}
	return f.vm.safeCallFunc(f, args)
}
