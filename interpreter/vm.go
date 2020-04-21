package interpreter

import (
	"fmt"
	"math"
	"strings"

	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
	"github.com/zxh0/wasm.go/validator"
)

var _ instance.Instance = (*vm)(nil)

type vm struct {
	operandStack
	controlStack

	module  binary.Module
	memory  instance.Memory
	table   instance.Table
	globals []instance.Global
	funcs   []vmFunc

	local0Idx uint32
	debug     bool
}

func NewInstance(m binary.Module, instances instance.Map) (inst instance.Instance, err error) {
	if err := validator.Validate(m); err != nil {
		return nil, err
	}

	defer func() {
		if _err := recover(); _err != nil {
			switch x := _err.(type) {
			case error:
				err = x
			default:
				panic(err)
			}
		}
	}()

	vm := &vm{module: m, debug: false}
	vm.linkImports(instances)
	vm.initFuncs()
	vm.initTableAndMem()
	vm.initGlobals()
	vm.execStartFunc()
	inst = vm
	return
}

/* linking */

func (vm *vm) linkImports(instances instance.Map) {
	for _, imp := range vm.module.ImportSec {
		if m := instances[imp.Module]; m == nil {
			panic(fmt.Errorf("module not found: " + imp.Module))
		} else {
			vm.linkImport(m, imp)
		}
	}
}
func (vm *vm) linkImport(m instance.Instance, imp binary.Import) {
	exported := m.Get(imp.Name)
	if exported == nil {
		panic(fmt.Errorf("unknown import: %s.%s",
			imp.Module, imp.Name))
	}

	typeMatched := false
	switch x := exported.(type) {
	case instance.Function:
		if imp.Desc.Tag == binary.ImportTagFunc {
			expectedFT := vm.module.TypeSec[imp.Desc.FuncType]
			typeMatched = isFuncTypeMatch(expectedFT, x.Type())
			vm.funcs = append(vm.funcs, newExternalFunc(vm, expectedFT, x))
		}
	case instance.Table:
		if imp.Desc.Tag == binary.ImportTagTable {
			typeMatched = isLimitsMatch(imp.Desc.Table.Limits, x.Type().Limits)
			vm.table = x
		}
	case instance.Memory:
		if imp.Desc.Tag == binary.ImportTagMem {
			typeMatched = isLimitsMatch(imp.Desc.Mem, x.Type())
			vm.memory = x
		}
	case instance.Global:
		if imp.Desc.Tag == binary.ImportTagGlobal {
			typeMatched = isGlobalTypeMatch(imp.Desc.Global, x.Type())
			vm.globals = append(vm.globals, x)
		}
	}

	if !typeMatched {
		panic(fmt.Errorf("incompatible import type: %s.%s",
			imp.Module, imp.Name))
	}
}

/* init */

func (vm *vm) initFuncs() {
	for i, sigIdx := range vm.module.FuncSec {
		sig := vm.module.TypeSec[sigIdx]
		code := vm.module.CodeSec[i]
		vm.funcs = append(vm.funcs, newInternalFunc(vm, sig, code))
	}
}

func (vm *vm) initTableAndMem() {
	if len(vm.module.TableSec) > 0 {
		vm.table = newTable(vm.module.TableSec[0])
	}
	if len(vm.module.MemSec) > 0 {
		vm.memory = newMemory(vm.module.MemSec[0])
	}
	elemOffsets := vm.calcElemOffsets()
	dataOffsets := vm.calcDataOffsets()
	vm.initTable(elemOffsets)
	vm.initMemory(dataOffsets)
}
func (vm *vm) calcElemOffsets() []uint32 {
	offsets := make([]uint32, len(vm.module.ElemSec))
	for i, elem := range vm.module.ElemSec {
		vm.execConstExpr(elem.Offset)
		offset := vm.popU32()
		dataLen := len(elem.Init)
		upperBound := vm.table.Type().Limits.Min
		if offset > 0 || dataLen > 0 {
			if uint64(offset)+uint64(dataLen) > uint64(upperBound) {
				panic(fmt.Errorf("elements segment does not fit"))
			}
		}
		offsets[i] = offset
	}
	return offsets
}
func (vm *vm) calcDataOffsets() []uint64 {
	offsets := make([]uint64, len(vm.module.DataSec))
	for i, data := range vm.module.DataSec {
		vm.execConstExpr(data.Offset)
		offset := uint64(vm.popU32())
		dataLen := uint64(len(data.Init))
		upperBound := uint64(vm.memory.Type().Min) * binary.PageSize
		if offset > 0 || dataLen > 0 {
			if offset+dataLen > upperBound {
				panic(fmt.Errorf("data segment does not fit"))
			}
		}
		offsets[i] = offset
	}
	return offsets
}
func (vm *vm) initTable(offsets []uint32) {
	for i, elem := range vm.module.ElemSec {
		for j, fIdx := range elem.Init {
			offset := offsets[i] + uint32(j)
			f := vm.funcs[fIdx]
			vm.table.SetElem(offset, f)
		}
	}
}
func (vm *vm) initMemory(offsets []uint64) {
	for i, data := range vm.module.DataSec {
		vm.memory.Write(offsets[i], data.Init)
	}
}

func (vm *vm) initGlobals() {
	for _, g := range vm.module.GlobalSec {
		initVal := uint64(0)
		if g.Init != nil {
			vm.execConstExpr(g.Init)
			initVal = vm.popU64()
		}
		g := newGlobal(g.Type, initVal)
		vm.globals = append(vm.globals, g)
	}
}

func (vm *vm) execConstExpr(expr []binary.Instruction) {
	for _, instr := range expr {
		vm.execInstr(instr)
	}
}
func (vm *vm) execStartFunc() {
	if vm.module.StartSec != nil {
		idx := *vm.module.StartSec
		vm.callFunc(vm.funcs[idx], nil)
	}
}

/* block stack */

func (vm *vm) enterBlock(opcode byte,
	bt binary.FuncType, instrs []binary.Instruction) {

	bp := vm.stackSize() - len(bt.ParamTypes)
	cf := newControlFrame(opcode, bt, instrs, bp)
	vm.pushControlFrame(cf)
	if opcode == binary.Call {
		vm.local0Idx = uint32(bp)
	}
}
func (vm *vm) exitBlock() {
	cf := vm.popControlFrame()
	vm.clearBlock(cf)
}
func (vm *vm) clearBlock(cf *controlFrame) {
	results := vm.popU64s(len(cf.bt.ResultTypes))
	vm.popU64s(vm.stackSize() - cf.bp)
	vm.pushU64s(results)
	if cf.opcode == binary.Call && vm.controlDepth() > 0 {
		lastCallFrame, _ := vm.topCallFrame()
		vm.local0Idx = uint32(lastCallFrame.bp)
	}
}
func (vm *vm) resetBlock(cf *controlFrame) {
	results := vm.popU64s(len(cf.bt.ParamTypes))
	vm.popU64s(vm.stackSize() - cf.bp)
	vm.pushU64s(results)
}

/* func call */

func (vm *vm) reset() {
	vm.operandStack.reset()
	vm.controlStack.reset()
}

func (vm *vm) safeCallFunc(f vmFunc,
	args []interface{}) (results []interface{}, err error) {

	defer func() {
		if _err := recover(); _err != nil {
			switch x := _err.(type) {
			case error:
				vm.reset()
				err = x
			default:
				panic(err)
			}
		}
	}()

	if vm.debug {
		fmt.Printf("safe call! %v\n", f) // TODO
	}

	results = vm.callFunc(f, args)
	return
}

func (vm *vm) callFunc(f vmFunc, args []interface{}) []interface{} {
	vm.pushArgs(f._type, args)
	callFunc(vm, f)
	if f.goFunc == nil {
		vm.loop()
	}
	return vm.popResults(f._type)
}

func (vm *vm) loop() {
	depth := vm.controlDepth()
	for vm.controlDepth() >= depth {
		cf := vm.topControlFrame()
		if cf.pc == len(cf.instrs) {
			vm.exitBlock()
		} else {
			instr := cf.instrs[cf.pc]
			cf.pc++
			vm.execInstr(instr)
		}
	}
}

func (vm *vm) execInstr(instr binary.Instruction) {
	vm.logInstr(instr)
	instrTable[instr.Opcode](vm, instr.Args)
}

func (vm *vm) logInstr(instr binary.Instruction) {
	if vm.debug {
		fmt.Print(strings.Repeat(">", vm.controlDepth()))
		if instr.Opcode != binary.Call {
			fmt.Printf("%s %v\n", instr.GetOpname(), instr.Args)
		} else {
			f := vm.funcs[instr.Args.(uint32)]
			fmt.Printf("call func#%d(", instr.Args)
			if n := len(f._type.ParamTypes); n > 0 {
				stack := vm.operandStack.slots
				fmt.Print(stack[len(stack)-n:])
			}
			fmt.Println(")")
		}
	}
}

func (vm *vm) pushArgs(ft binary.FuncType, args []interface{}) {
	if len(ft.ParamTypes) != len(args) {
		panic(fmt.Errorf("param count: %d, arg count: %d",
			len(ft.ParamTypes), len(args)))
	}
	for i, vt := range ft.ParamTypes {
		switch vt {
		case binary.ValTypeI32:
			vm.pushS32(args[i].(int32))
		case binary.ValTypeI64:
			vm.pushS64(args[i].(int64))
		case binary.ValTypeF32:
			vm.pushF32(args[i].(float32))
		case binary.ValTypeF64:
			vm.pushF64(args[i].(float64))
		default:
			panic("unreachable")
		}
	}
}
func (vm *vm) popResults(ft binary.FuncType) []interface{} {
	results := make([]interface{}, len(ft.ResultTypes))
	for n := len(ft.ResultTypes) - 1; n >= 0; n-- {
		switch ft.ResultTypes[n] {
		case binary.ValTypeI32:
			results[n] = vm.popS32()
		case binary.ValTypeI64:
			results[n] = vm.popS64()
		case binary.ValTypeF32:
			results[n] = vm.popF32()
		case binary.ValTypeF64:
			results[n] = vm.popF64()
		default:
			panic("unreachable")
		}
	}
	return results
}

/* instance.Instance */

func (vm *vm) Get(name string) interface{} {
	for _, exp := range vm.module.ExportSec {
		if exp.Name == name {
			idx := exp.Desc.Idx
			switch exp.Desc.Tag {
			case binary.ExportTagFunc:
				return vm.funcs[idx]
			case binary.ExportTagTable:
				return vm.table
			case binary.ExportTagMem:
				return vm.memory
			case binary.ExportTagGlobal:
				return vm.globals[idx]
			}
		}
	}
	return nil
}

func (vm vm) GetGlobalValue(name string) (interface{}, error) {
	for _, exp := range vm.module.ExportSec {
		if exp.Name == name && exp.Desc.Tag == binary.ExportTagGlobal {
			g := vm.globals[exp.Desc.Idx]
			switch g.Type().ValType {
			case binary.ValTypeI32:
				return int32(uint32(g.Get())), nil
			case binary.ValTypeI64:
				return int64(g.Get()), nil
			case binary.ValTypeF32:
				return math.Float32frombits(uint32(g.Get())), nil
			case binary.ValTypeF64:
				return math.Float64frombits(g.Get()), nil
			default:
				panic("unreachable")
			}
		}
	}
	return nil, fmt.Errorf("global not found: " + name)
}

func (vm *vm) CallFunc(name string, args ...interface{}) ([]interface{}, error) {
	fIdx, ok := vm.getFunc(name) // TODO
	if !ok {
		return nil, fmt.Errorf("function not found: " + name)
	}

	return vm.funcs[fIdx].Call(args...)
}

func (vm *vm) getFunc(name string) (uint32, bool) {
	for _, exp := range vm.module.ExportSec {
		if exp.Name == name && exp.Desc.Tag == binary.ExportTagFunc {
			return exp.Desc.Idx, true
		}
	}
	return 0, false
}

/* helpers */

func isFuncTypeMatch(expected, actual binary.FuncType) bool {
	return fmt.Sprintf("%s", expected) == fmt.Sprintf("%s", actual)
}
func isGlobalTypeMatch(expected, actual binary.GlobalType) bool {
	return actual.ValType == expected.ValType &&
		actual.Mut == expected.Mut
}
func isLimitsMatch(expected, actual binary.Limits) bool {
	return actual.Min >= expected.Min &&
		(expected.Max == 0 || actual.Max > 0 && actual.Max <= expected.Max)
}
