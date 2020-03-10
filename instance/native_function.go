package instance

import (
	"github.com/zxh0/wasm.go/binary"
)

var _ Function = (*nativeFunction)(nil)

type nativeFunction struct {
	t binary.FuncType
	f GoFunc
}

func (nf nativeFunction) Type() binary.FuncType {
	return nf.t
}
func (nf nativeFunction) Call(args ...interface{}) (interface{}, error) {
	return nf.f(args...)
}
