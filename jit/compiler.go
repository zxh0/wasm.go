// +build jit

package jit

import (
	"fmt"

	"github.com/tinygo-org/go-llvm"
	"github.com/zxh0/wasm.go/binary"
)

type compiler struct {
	llvmMod llvm.Module
	builder llvm.Builder
}

func newCompiler() compiler {
	return compiler{
		llvmMod: llvm.NewModule("jit"),
		builder: llvm.NewBuilder(),
	}
}

func Compile(module binary.Module) {
	newCompiler().compileModule(module)
}

func (c compiler) compileModule(m binary.Module) {
	for i, ftIdx := range m.FuncSec {
		c.compileCode(i, m.TypeSec[ftIdx], m.CodeSec[i])
	}
	fmt.Println(c.llvmMod.String())
}

func (c compiler) compileCode(idx int, ft binary.FuncType, code binary.Code) {
	name := fmt.Sprintf("f%d", idx)
	llvm.AddFunction(c.llvmMod, name, funcType2LLVM(ft))
}

func funcType2LLVM(ft binary.FuncType) llvm.Type {
	var returnType llvm.Type
	if len(ft.ResultTypes) == 0 {
		returnType = llvm.VoidType()
	} else {
		returnType = valType2LLVM(ft.ResultTypes[0])
	}
	paramTypes := valTypes2LLVM(ft.ParamTypes)
	return llvm.FunctionType(returnType, paramTypes, false)
}

func valType2LLVM(vt binary.ValType) llvm.Type {
	switch vt {
	case binary.ValTypeI32:
		return llvm.Int32Type()
	case binary.ValTypeI64:
		return llvm.Int64Type()
	case binary.ValTypeF32:
		return llvm.FloatType()
	case binary.ValTypeF64:
		return llvm.DoubleType()
	default:
		panic("unreachable")
	}
}

func valTypes2LLVM(vts []binary.ValType) []llvm.Type {
	llvmTypes := make([]llvm.Type, len(vts))
	for i, vt := range vts {
		llvmTypes[i] = valType2LLVM(vt)
	}
	return llvmTypes
}
