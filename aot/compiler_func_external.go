package aot

import "github.com/zxh0/wasm.go/binary"

type externalFuncCompiler struct {
	funcCompiler
}

func newExternalFuncCompiler() *externalFuncCompiler {
	return &externalFuncCompiler{newFuncCompiler()}
}

func (c *externalFuncCompiler) compile(idx int, ft binary.FuncType) string {
	c.printf("func (m *aotModule) f%d(", idx)
	c.genParams(len(ft.ParamTypes))
	c.print(")")
	c.genResults(len(ft.ResultTypes))
	c.print(" {\n")
	c.genFuncBody(idx, ft)
	c.println("}")
	return c.sb.String()
}

func (c *externalFuncCompiler) genFuncBody(idx int, ft binary.FuncType) {
	c.printIf(len(ft.ResultTypes) > 0,
		"	r, err := ",
		"	_, err := ")
	c.printf("m.importedFuncs[%d].Call(", idx)
	for i, vt := range ft.ParamTypes {
		c.printIf(i > 0, ", ", "")
		switch vt {
		case binary.ValTypeI32:
			c.printf("int32(l%d)", i)
		case binary.ValTypeI64:
			c.printf("int64(l%d)", i)
		case binary.ValTypeF32:
			c.printf("f32(l%d)", i)
		case binary.ValTypeF64:
			c.printf("f64(l%d)", i)
		}
	}
	c.println(")")
	c.println("	if err != nil {} // TODO")
	if len(ft.ResultTypes) > 0 {
		switch ft.ResultTypes[0] {
		case binary.ValTypeI32:
			c.println("return uint32(r.(int32))")
		case binary.ValTypeI64:
			c.println("return uint64(r.(int64))")
		case binary.ValTypeF32:
			c.println("return u32(r.(float32))")
		case binary.ValTypeF64:
			c.println("return u64(r.(float64))")
		}
	}
}
