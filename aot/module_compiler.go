package aot

import "github.com/zxh0/wasm.go/binary"

type moduleCompiler struct {
	printer
	module binary.Module
}

func (c *moduleCompiler) compile() {
	c.genModule()
	c.genNew()
	c.genUtils()
	c.println("")
	m := c.module
	importedFuncCount := getImportedFuncCount(c.module)
	for i, ftIdx := range m.FuncSec {
		fc := newFuncCompiler(c.module)
		fIdx := importedFuncCount + i
		ft := m.TypeSec[ftIdx]
		code := m.CodeSec[i]
		c.println(fc.compile(fIdx, ft, code))
	}
}

func (c *moduleCompiler) genModule() {
	c.print(`
type aotModule struct {
	memory  []byte
	globals []uint64
}
`)
}

func (c *moduleCompiler) genNew() {
	memPageMin := getMemPageMin(c.module)
	globalCount := len(c.module.GlobalSec)
	c.printf(`
func New() *aotModule {
	return &aotModule{
		memory:  make([]byte, %d),
		globals: make([]uint64, %d),
	}
}
`, memPageMin*binary.PageSize, globalCount)
}

func (c *moduleCompiler) genUtils() {
	c.print(`
// utils
func b2i(b bool) uint64 {
	if b { return 1 } else { return 0 }
}
func f32(i uint64) float32 {
	return math.Float32frombits(uint32(i))
}
func u32(f float32) uint64 {
	return uint64(math.Float32bits(f))
}
func f64(i uint64) float64 {
	return math.Float64frombits(i)
}
func u64(f float32) uint64 {
	return math.Float64bits(f)
}
`)
}
