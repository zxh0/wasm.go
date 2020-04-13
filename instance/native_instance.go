package instance

import (
	"github.com/zxh0/wasm.go/binary"
)

var _ Instance = (*NativeInstance)(nil)

type NativeInstance struct {
	exported map[string]interface{}
}

func NewNativeInstance() *NativeInstance {
	return &NativeInstance{
		exported: map[string]interface{}{},
	}
}

func (n *NativeInstance) RegisterFunc(name string,
	f GoFunc, paramsAndResult ...binary.ValType) {

	ft := binary.FuncType{}
	if len(paramsAndResult) > 0 {
		ft.ParamTypes = paramsAndResult[:len(paramsAndResult)-1]
		rt := paramsAndResult[len(paramsAndResult)-1]
		if rt != binary.NoVal {
			ft.ResultTypes = []binary.ValType{rt}
		}
	}

	n.exported[name] = nativeFunction{t: ft, f: f}
}

func (n *NativeInstance) Register(name string, x interface{}) {
	n.exported[name] = x
}

func (n *NativeInstance) Get(name string) interface{} {
	return n.exported[name]
}

func (n *NativeInstance) CallFunc(name string, args ...interface{}) ([]interface{}, error) {
	return n.exported[name].(Function).Call(args...) // TODO
}

func (n *NativeInstance) GetGlobalValue(name string) (interface{}, error) {
	return n.exported[name].(Global).Get(), nil // TODO
}
