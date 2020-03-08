package interpreter

import (
	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
)

var _ instance.Memory = (*memory)(nil)

type memory struct {
	_type binary.MemType
	data  []byte
}

func NewMemory(mt binary.MemType) instance.Memory {
	return newMemory(mt)
}

func newMemory(mt binary.MemType) *memory {
	return &memory{
		_type: mt,
		data:  make([]byte, mt.Min*binary.PageSize),
	}
}

func (mem *memory) Type() binary.MemType {
	return mem._type
}

func (mem *memory) Size() uint32 {
	return uint32(len(mem.data) / binary.PageSize)
}
func (mem *memory) Grow(n uint32) uint32 {
	curPageCount := mem.Size()
	if n == 0 {
		return curPageCount
	}

	maxPageCount := uint32(binary.MaxPageCount)
	if max := mem._type.Max; max > 0 {
		maxPageCount = max
	}
	if curPageCount+n > maxPageCount {
		return 0xFFFFFFFF // -1
	}

	newData := make([]byte, (curPageCount+n)*binary.PageSize)
	copy(newData, mem.data)
	mem.data = newData
	return curPageCount
}

func (mem *memory) Read(offset uint64, buf []byte) {
	mem.checkOffset(offset, len(buf))
	copy(buf, mem.data[offset:])
}
func (mem *memory) Write(offset uint64, data []byte) {
	mem.checkOffset(offset, len(data))
	copy(mem.data[offset:], data)
}

func (mem *memory) checkOffset(offset uint64, length int) {
	if int64(len(mem.data)-length) < int64(offset) {
		panic("out of bounds memory access")
	}
}
