package instance

import "github.com/zxh0/wasm.go/binary"

type Map = map[string]Instance
type GoFunc = func(args ...interface{}) ([]interface{}, error)

type Instance interface {
	Get(name string) interface{}
	CallFunc(name string, args ...interface{}) ([]interface{}, error)
	GetGlobalValue(name string) (interface{}, error)
}

type Function interface {
	Type() binary.FuncType
	Call(args ...interface{}) ([]interface{}, error)
}

type Table interface {
	Type() binary.TableType
	Size() uint32
	Grow(n uint32)
	GetElem(idx uint32) Function
	SetElem(idx uint32, elem Function)
}

type Memory interface {
	Type() binary.MemType
	Size() uint32 // page count
	Grow(n uint32) uint32
	Read(offset uint64, buf []byte)
	Write(offset uint64, buf []byte)
}

type Global interface {
	Type() binary.GlobalType
	Get() uint64
	Set(val uint64)
}
