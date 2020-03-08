package interpreter

import (
	"github.com/zxh0/wasm.go/binary"
)

func unreachable(vm *vm, _ interface{}) {
	panic("unreachable") // TODO
}

func nop(vm *vm, _ interface{}) {
	// do nothing
}

func block(vm *vm, args interface{}) {
	blockArgs := args.(binary.BlockArgs)
	vm.enterBlock(blockArgs.Instrs, blockArgs.RT, btBlock, 0)
}

func loop(vm *vm, args interface{}) {
	blockArgs := args.(binary.BlockArgs)
	vm.enterBlock(blockArgs.Instrs, blockArgs.RT, btLoop, 0)
}

func _if(vm *vm, args interface{}) {
	ifArgs := args.(binary.IfArgs)
	if vm.popBool() {
		vm.enterBlock(ifArgs.Instrs1, ifArgs.RT, btBlock, 0)
	} else {
		vm.enterBlock(ifArgs.Instrs2, ifArgs.RT, btBlock, 0)
	}
}

func br(vm *vm, args interface{}) {
	labelIdx := int(args.(uint32))
	for i := 0; i < labelIdx; i++ {
		vm.popBlockFrame()
	}
	if bf := vm.topBlockFrame(); bf.bt != btLoop {
		vm.exitBlock()
	} else {
		vm.clearBlock(bf)
		bf.pc = 0
	}
}

func brIf(vm *vm, args interface{}) {
	if vm.popBool() {
		br(vm, args)
	}
}

func brTable(vm *vm, args interface{}) {
	brTableArgs := args.(binary.BrTableArgs)
	n := int(vm.popU32())
	if n < len(brTableArgs.Labels) {
		br(vm, brTableArgs.Labels[n])
	} else {
		br(vm, brTableArgs.Default)
	}
}

func _return(vm *vm, _ interface{}) {
	var bf *blockFrame
	for {
		bf = vm.popBlockFrame()
		if bf.bt == btFunc {
			break
		}
	}
	vm.clearBlock(bf)
}

func call(vm *vm, args interface{}) {
	f := vm.funcs[args.(uint32)]
	callFunc(vm, f)
}

func callFunc(vm *vm, f vmFunc) {
	if f.imported != nil {
		callExternalFunc(vm, f)
	} else {
		callInternalFunc(vm, f)
	}
}

func callExternalFunc(vm *vm, f vmFunc) {
	args := popArgs(vm, f._type)
	result, err := f.imported.Call(args...)
	if err != nil {
		panic(err)
	}
	pushResult(vm, f._type, result)
}

func popArgs(vm *vm, sig binary.FuncType) []interface{} {
	paramCount := len(sig.ParamTypes)
	args := make([]interface{}, paramCount)
	for i := paramCount - 1; i >= 0; i-- {
		switch sig.ParamTypes[i] {
		case binary.ValTypeI32:
			args[i] = vm.popS32()
		case binary.ValTypeI64:
			args[i] = vm.popS64()
		case binary.ValTypeF32:
			args[i] = vm.popF32()
		case binary.ValTypeF64:
			args[i] = vm.popF64()
		}
	}
	return args
}

func pushResult(vm *vm, sig binary.FuncType, result interface{}) {
	if len(sig.ResultTypes) > 0 {
		switch sig.ResultTypes[0] {
		case binary.ValTypeI32:
			vm.pushS32(result.(int32))
		case binary.ValTypeI64:
			vm.pushS64(result.(int64))
		case binary.ValTypeF32:
			vm.pushF32(result.(float32))
		case binary.ValTypeF64:
			vm.pushF64(result.(float64))
		}
	}
}

/*
operand stack:

+~~~~~~~~~~~~~~~+
|               |
+---------------+
|     stack     |
+---------------+
|     locals    |
+---------------+
|     params    |
+---------------+
|  ............ |
*/
func callInternalFunc(vm *vm, f vmFunc) {
	// alloc locals
	localCount := f.code.GetLocalCount()
	for i := 0; i < localCount; i++ {
		vm.pushU64(0)
	}

	paramsCount := len(f._type.ParamTypes)
	vm.enterBlock(f.code.Expr, f._type.ResultTypes, btFunc, localCount+paramsCount)
}

func callIndirect(vm *vm, args interface{}) {
	typeIdx := args.(uint32)
	ft := vm.module.TypeSec[typeIdx]

	i := vm.popU32()
	if i >= vm.table.Size() {
		panic("undefined element") // TODO
	}

	f := vm.table.GetElem(i)
	if f.Type().GetSignature() != ft.GetSignature() {
		panic("indirect call type mismatch") // TODO
	}

	// optimize internal func call
	if _f, ok := f.(vmFunc); ok {
		if _f.imported == nil && _f.vm == vm {
			callInternalFunc(vm, _f)
			return
		}
	}

	fcArgs := popArgs(vm, ft)
	result, err := f.Call(fcArgs...)
	if err != nil {
		panic(err)
	}
	pushResult(vm, ft, result)
}
