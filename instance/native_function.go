package instance

import (
	"errors"
	"reflect"

	"github.com/zxh0/wasm.go/binary"
)

var _ Function = (*nativeFunction)(nil)

type GoFunc = func(args []WasmVal) ([]WasmVal, error)

type nativeFunction struct {
	t binary.FuncType
	f GoFunc
}

func (nf nativeFunction) Type() binary.FuncType {
	return nf.t
}
func (nf nativeFunction) Call(args ...WasmVal) ([]WasmVal, error) {
	return nf.f(args)
}

func wrapNativeFunc(nf interface{}) (Function, error) {
	ft, err := getNativeFuncType(nf)
	if err != nil {
		return nil, err
	}

	f := func(args []WasmVal) ([]WasmVal, error) {
		return callNativeFunc(ft, nf, args...)
	}

	return nativeFunction{ft, f}, nil
}

func callNativeFunc(ft binary.FuncType,
	nf interface{}, args ...WasmVal) ([]WasmVal, error) {

	paramCount := len(ft.ParamTypes)
	resultCount := len(ft.ResultTypes)
	if paramCount != len(args) {
		return nil, errors.New("wrong number of args")
	}

	typeOK := true
	in := make([]reflect.Value, paramCount)
	for i, paramVt := range ft.ParamTypes {
		in[i] = reflect.ValueOf(args[i])
		argVt, err := getNativeValType(in[i].Kind())
		if err != nil || argVt != paramVt {
			typeOK = false
			break
		}
	}
	if !typeOK {
		return nil, errors.New("arg type mismatch")
	}

	out := reflect.ValueOf(nf).Call(in)
	if len(out) != resultCount {
		return nil, errors.New("wrong number of results")
	}
	results := make([]interface{}, resultCount)
	for i, r := range out {
		rt, err := getNativeValType(r.Kind())
		if err != nil || rt != ft.ResultTypes[i] {
			return nil, errors.New("result type mismatch")
		}
		results[i] = r.Interface()
	}

	return results, nil
}

func getNativeFuncType(nf interface{}) (ft binary.FuncType, err error) {
	nfType := reflect.TypeOf(nf)
	if nfType.Kind() != reflect.Func {
		err = errors.New("not a function")
		return
	}

	var vt binary.ValType
	for i := 0; i < nfType.NumIn(); i++ {
		if vt, err = getNativeValType(nfType.In(i).Kind()); err != nil {
			return
		}
		ft.ParamTypes = append(ft.ParamTypes, vt)
	}
	for i := 0; i < nfType.NumOut(); i++ {
		if vt, err = getNativeValType(nfType.Out(i).Kind()); err != nil {
			return
		}
		ft.ResultTypes = append(ft.ResultTypes, vt)
	}
	return
}

func getNativeValType(kind reflect.Kind) (binary.ValType, error) {
	switch kind {
	case reflect.Int32:
		return binary.ValTypeI32, nil
	case reflect.Int64:
		return binary.ValTypeI64, nil
	case reflect.Float32:
		return binary.ValTypeF32, nil
	case reflect.Float64:
		return binary.ValTypeF64, nil
	default:
		return 0, errors.New("unsupported type: " + kind.String())
	}
}
