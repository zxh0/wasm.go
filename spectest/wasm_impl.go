package spectest

import (
	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
	"github.com/zxh0/wasm.go/interpreter"
	"github.com/zxh0/wasm.go/validator"
)

var _ WasmImpl = (*WasmInterpreter)(nil)

type WasmImpl interface {
	Validate(m binary.Module) error
	Instantiate(m binary.Module, instances instance.Map) (instance.Module, error)
	InstantiateBin(data []byte, instances instance.Map) (instance.Module, error)
}

type WasmInterpreter struct {
}

func (WasmInterpreter) Validate(m binary.Module) error {
	return validator.Validate(m)
}

func (WasmInterpreter) Instantiate(
	m binary.Module, instances instance.Map) (instance.Module, error) {

	return interpreter.New(m, instances)
}

func (WasmInterpreter) InstantiateBin(
	data []byte, instances instance.Map) (instance.Module, error) {

	m, err := binary.Decode(data)
	if err != nil {
		return nil, err
	}
	return interpreter.New(m, instances)
}
