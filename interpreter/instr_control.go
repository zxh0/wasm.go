package interpreter

import (
	"github.com/zxh0/wasm.go/binary"
)

func unreachable(vm *vm, _ interface{}) {
	panic(errTrap)
}

func nop(vm *vm, _ interface{}) {
	// do nothing
}

func block(vm *vm, args interface{}) {
	blockArgs := args.(binary.BlockArgs)
	bt := vm.module.GetBlockType(blockArgs.BT)
	vm.enterBlock(binary.Block, bt, blockArgs.Instrs)
}

func loop(vm *vm, args interface{}) {
	blockArgs := args.(binary.BlockArgs)
	bt := vm.module.GetBlockType(blockArgs.BT)
	vm.enterBlock(binary.Loop, bt, blockArgs.Instrs)
}

func _if(vm *vm, args interface{}) {
	ifArgs := args.(binary.IfArgs)
	bt := vm.module.GetBlockType(ifArgs.BT)
	if vm.popBool() {
		vm.enterBlock(binary.If, bt, ifArgs.Instrs1)
	} else {
		vm.enterBlock(binary.If, bt, ifArgs.Instrs2)
	}
}

func br(vm *vm, args interface{}) {
	labelIdx := int(args.(uint32))
	for i := 0; i < labelIdx; i++ {
		vm.popControlFrame()
	}
	if cf := vm.topControlFrame(); cf.opcode != binary.Loop {
		vm.exitBlock()
	} else {
		vm.resetBlock(cf)
		cf.pc = 0
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
	var cf *controlFrame
	for {
		cf = vm.popControlFrame()
		if cf.opcode == binary.Call {
			break
		}
	}
	vm.clearBlock(cf)
}

func call(vm *vm, args interface{}) {
	f := vm.funcs[args.(uint32)]
	callFunc(vm, f)
}

func callFunc(vm *vm, f vmFunc) {
	if f.goFunc != nil {
		callExternalFunc(vm, f)
	} else {
		callInternalFunc(vm, f)
	}
}

func callExternalFunc(vm *vm, f vmFunc) {
	args := popArgs(vm, f._type)
	results, err := f.goFunc.Call(args...)
	if err != nil {
		panic(err)
	}
	pushResults(vm, f._type, results)
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

func pushResults(vm *vm, sig binary.FuncType, results []interface{}) {
	if len(sig.ResultTypes) != len(results) {
		panic("TODO")
	}
	for _, result := range results {
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
	vm.enterBlock(binary.Call, f._type, f.code.Expr)

	// alloc locals
	localCount := int(f.code.GetLocalCount())
	for i := 0; i < localCount; i++ {
		vm.pushU64(0)
	}
}

func callIndirect(vm *vm, args interface{}) {
	typeIdx := args.(uint32)
	ft := vm.module.TypeSec[typeIdx]

	i := vm.popU32()
	if i >= vm.table.Size() {
		panic(errUndefinedElem)
	}

	f := vm.table.GetElem(i)
	if f.Type().GetSignature() != ft.GetSignature() {
		panic(errTypeMismatch)
	}

	// optimize internal func call
	if _f, ok := f.(vmFunc); ok {
		if _f.goFunc == nil && _f.vm == vm {
			callInternalFunc(vm, _f)
			return
		}
	}

	fcArgs := popArgs(vm, ft)
	results, err := f.Call(fcArgs...)
	if err != nil {
		panic(err)
	}
	pushResults(vm, ft, results)
}
