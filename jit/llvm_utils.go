// +build jit

package jit

import (
	"github.com/tinygo-org/go-llvm"
	"github.com/zxh0/wasm.go/binary"
)

func constI32(n int32) llvm.Value {
	return llvm.ConstInt(llvm.Int32Type(), uint64(n), true)
}
func constI64(n int64) llvm.Value {
	return llvm.ConstInt(llvm.Int64Type(), uint64(n), true)
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

func findCtrlInstr(instrs []binary.Instruction) int {
	for i, instr := range instrs {
		switch instr.Opcode {
		case binary.Block, binary.Loop, binary.If,
			binary.Br, binary.BrIf, binary.BrTable, binary.Return:
			return i
		}
	}
	return -1
}
