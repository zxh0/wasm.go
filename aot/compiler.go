package aot

import (
	"fmt"
	"strings"

	"github.com/zxh0/wasm.go/binary"
)

func Compile(module binary.Module) {
	c := &moduleCompiler{
		printer:    newPrinter(),
		moduleInfo: newModuleInfo(module),
	}
	c.compile()
	fmt.Println(c.sb.String())
}

func CompileFunc(module binary.Module, idx int) string {
	moduleInfo := newModuleInfo(module)
	ftIdx := moduleInfo.module.FuncSec[idx-len(moduleInfo.importedFuncs)]
	ft := moduleInfo.module.TypeSec[ftIdx]
	code := moduleInfo.module.CodeSec[idx-len(moduleInfo.importedFuncs)]
	fc := newInternalFuncCompiler(moduleInfo)
	s := fc.compile(idx, ft, code)
	return wrapCompiledFunc(s, idx, ft)
}

func wrapCompiledFunc(s string, idx int, ft binary.FuncType) string {
	//s = strings.Replace(s, " (m *aotModule)", "", 1)
	s = strings.Replace(s, fmt.Sprintf("(m *aotModule) f%d", idx), "Call", 1)
	s = "package main\n\n" + s
	s += "\n"
	s += "func b2i(b bool) uint64 { if b { return 1 } else { return 0 } }\n\n"
	//s += genFuncWrapper(idx, ft)
	return s
}

func genFuncWrapper(idx int, ft binary.FuncType) string {
	p := newPrinter()
	p.println("func Call(args []uint64) []uint64 {")
	p.print("\t")
	if len(ft.ResultTypes) > 0 {
		for i := range ft.ResultTypes {
			p.printIf(i > 0, ", ", "")
			p.printf("r%d", i)
		}
		p.print(" := ")
	}
	p.printf("f%d(", idx)
	for i := range ft.ParamTypes {
		p.printIf(i > 0, ", ", "")
		p.printf("args[%d]", i)
	}
	p.println(")")
	if len(ft.ResultTypes) > 0 {
		p.print("\treturn []uint64{")
		if len(ft.ResultTypes) > 0 {
			for i := range ft.ResultTypes {
				p.printIf(i > 0, ", ", "")
				p.printf("r%d", i)
			}
		}
		p.println("}")
	} else {
		p.println("\treturn nil")
	}
	p.println("}")
	return p.String()
}
