package aot

import "github.com/zxh0/wasm.go/binary"

type exportedFuncCompiler struct {
	funcCompiler
	importedFuncCount int
}

func newExportedFuncCompiler(importedFuncCount int) *exportedFuncCompiler {
	return &exportedFuncCompiler{
		funcCompiler:      newFuncCompiler(),
		importedFuncCount: importedFuncCount,
	}
}

func (c *exportedFuncCompiler) compile(expIdx, fIdx int, ft binary.FuncType) string {
	c.printf("func (m *aotModule) exported%d(args ...interface{}) (interface{}, error) {\n", expIdx)
	if fIdx < c.importedFuncCount {
		c.printf("	return m.f%d(args...)\n", fIdx)
	} else {
		c.print("	")
		c.printIf(len(ft.ResultTypes) > 0, "r := ", "")
		c.printf("m.f%d(", fIdx)
		for i, vt := range ft.ParamTypes {
			c.printIf(i > 0, ", ", "")
			switch vt {
			case binary.ValTypeI32:
				c.printf("uint64(args[%d].(int32))", i)
			case binary.ValTypeI64:
				c.printf("uint64(args[%d].(int64))", i)
			case binary.ValTypeF32:
				c.printf("u32(args[%d].(float32))", i)
			case binary.ValTypeF64:
				c.printf("u64(args[%d].(float64))", i)
			}
		}
		c.println(")")
		if len(ft.ResultTypes) > 0 {
			switch ft.ResultTypes[0] {
			case binary.ValTypeI32:
				c.println("	return int32(r), nil")
			case binary.ValTypeI64:
				c.println("	return int64(r), nil")
			case binary.ValTypeF32:
				c.println("	return f32(r), nil")
			case binary.ValTypeF64:
				c.println("	return f64(r), nil")
			}
		} else {
			c.println("	return nil, nil")
		}
	}
	c.println("}")
	return c.sb.String()
}
