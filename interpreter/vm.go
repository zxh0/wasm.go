package interpreter

import (
	"errors"
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
	blockStack

	module  binary.Module
	memory  instance.Memory
	table   instance.Table
	globals []instance.Global
	funcs   []vmFunc

	local0Idx uint32
	debug     bool
}

func NewInstance(m binary.Module, instances instance.Map) (instance.Instance, error) {
	if err := validator.Validate(m); err != nil {
		return nil, err
	}

	vm := &vm{module: m, debug: false}
	if err := vm.linkImports(instances); err != nil {
		return nil, err
	}

	vm.initFuncs()
	if err := vm.initTableAndMem(); err != nil {
		return nil, err
	}
	vm.initGlobals()
	if err := vm.execStartFunc(); err != nil {
		return nil, err
	}
	return vm, nil
}

/* linking */

func (vm *vm) linkImports(instances instance.Map) error {
	for _, imp := range vm.module.ImportSec {
		m := instances[imp.Module]
		if m == nil {
			return fmt.Errorf("module not found: " + imp.Module)
		}
		if err := vm.linkImport(m, imp); err != nil {
			return err
		}
	}
	return nil
}
func (vm *vm) linkImport(m instance.Instance, imp binary.Import) error {
	exported := m.Get(imp.Name)
	if exported == nil {
		return fmt.Errorf("unknown import: %s.%s",
			imp.Module, imp.Name)
	}

	typeMatched := false

	switch x := exported.(type) {
	case instance.Function:
		if imp.Desc.Tag == binary.ImportTagFunc {
			expectedFT := vm.module.TypeSec[imp.Desc.FuncType]
			if isFuncTypeMatch(expectedFT, x.Type()) {
				typeMatched = true
				vm.funcs = append(vm.funcs,
					newExternalFunc(vm, expectedFT, x))
			}
		}
	case instance.Table:
		if imp.Desc.Tag == binary.ImportTagTable {
			if isLimitsMatch(imp.Desc.Table.Limits, x.Type().Limits) {
				typeMatched = true
				vm.table = x
			}
		}
	case instance.Memory:
		if imp.Desc.Tag == binary.ImportTagMem {
			if isLimitsMatch(imp.Desc.Mem, x.Type()) {
				typeMatched = true
				vm.memory = x
			}
		}
	case instance.Global:
		if imp.Desc.Tag == binary.ImportTagGlobal {
			if isGlobalTypeMatch(imp.Desc.Global, x.Type()) {
				typeMatched = true
				vm.globals = append(vm.globals, x)
			}
		}
	}

	if !typeMatched {
		return fmt.Errorf("incompatible import type: %s.%s",
			imp.Module, imp.Name)
	}
	return nil
}

/* init */

func (vm *vm) initFuncs() {
	for i, sigIdx := range vm.module.FuncSec {
		sig := vm.module.TypeSec[sigIdx]
		code := vm.module.CodeSec[i]
		vm.funcs = append(vm.funcs, newInternalFunc(vm, sig, code))
	}
}

func (vm *vm) initTableAndMem() error {
	if len(vm.module.TableSec) > 0 {
		vm.table = newTable(vm.module.TableSec[0])
	}
	if len(vm.module.MemSec) > 0 {
		vm.memory = newMemory(vm.module.MemSec[0])
	}
	elemOffsets, err := vm.calcElemOffsets()
	if err != nil {
		return err
	}
	dataOffsets, err := vm.calcDataOffsets()
	if err != nil {
		return err
	}
	vm.initTable(elemOffsets)
	vm.initMemory(dataOffsets)
	return nil
}
func (vm *vm) calcElemOffsets() ([]uint32, error) {
	offsets := make([]uint32, len(vm.module.ElemSec))
	for i, elem := range vm.module.ElemSec {
		vm.execConstExpr(elem.Offset)
		offset := vm.popU32()
		dataLen := len(elem.Init)
		upperBound := vm.table.Type().Limits.Min
		if offset > 0 || dataLen > 0 {
			if uint64(offset)+uint64(dataLen) > uint64(upperBound) {
				return nil, fmt.Errorf("elements segment does not fit")
			}
		}
		// ok
		offsets[i] = offset
	}
	return offsets, nil
}
func (vm *vm) calcDataOffsets() ([]uint64, error) {
	offsets := make([]uint64, len(vm.module.DataSec))
	for i, data := range vm.module.DataSec {
		vm.execConstExpr(data.Offset)
		offset := uint64(vm.popU32())
		dataLen := uint64(len(data.Init))
		upperBound := uint64(vm.memory.Type().Min) * binary.PageSize
		if offset > 0 || dataLen > 0 {
			if offset+dataLen > upperBound {
				return nil, fmt.Errorf("data segment does not fit")
			}
		}
		// ok
		offsets[i] = offset
	}
	return offsets, nil
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
		if g.Expr != nil {
			vm.execConstExpr(g.Expr)
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
func (vm *vm) execStartFunc() error {
	if vm.module.StartSec == nil {
		return nil
	}
	idx := *vm.module.StartSec
	_, err := vm.safeCallFunc(vm.funcs[idx], nil)
	return err
}

/* block stack */

func (vm *vm) enterBlock(instrs []binary.Instruction,
	rt binary.BlockType, bt byte, localCount int) {

	bp := vm.stackSize() - localCount
	bf := newBlockFrame(instrs, rt, bt, bp)
	vm.pushBlockFrame(bf)
	if bt == btFunc {
		vm.local0Idx = uint32(bp)
	}
}
func (vm *vm) exitBlock() {
	bf := vm.popBlockFrame()
	vm.clearBlock(bf)
}
func (vm *vm) clearBlock(bf *blockFrame) {
	var result uint64
	if len(bf.rt) > 0 {
		result = vm.popU64()
	}
	for vm.stackSize() > bf.bp {
		vm.popU64()
	}
	if len(bf.rt) > 0 {
		vm.pushU64(result)
	}
	if bf.bt == btFunc && vm.blockDepth() > 0 {
		vm.local0Idx = uint32(vm.topFuncFrame().bp)
	}
}

/* func call */

func (vm *vm) reset() {
	vm.operandStack.reset()
	vm.blockStack.reset()
}

func (vm *vm) safeCallFunc(f vmFunc,
	args []interface{}) (result interface{}, err error) {

	defer func() {
		if _err := recover(); _err != nil {
			switch x := _err.(type) {
			case error:
				vm.reset()
				err = x
			case string:
				vm.reset()
				err = errors.New(x) // TODO
			default:
				panic(err)
			}
		}
	}()

	if vm.debug {
		fmt.Printf("safe call! %v\n", f) // TODO
	}

	result = vm.callFunc(f, args)
	return
}

func (vm *vm) callFunc(f vmFunc, args []interface{}) interface{} {
	vm.pushArgs(f._type, args)
	callFunc(vm, f)
	if f.imported == nil {
		vm.loop()
	}
	return vm.popResult(f._type)
}

func (vm *vm) loop() {
	depth := vm.blockDepth()
	for vm.blockDepth() >= depth {
		frame := vm.topBlockFrame()
		if frame.pc == len(frame.instrs) {
			vm.exitBlock()
		} else {
			instr := frame.instrs[frame.pc]
			frame.pc++
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
		fmt.Print(strings.Repeat(">", vm.blockDepth()))
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
func (vm *vm) popResult(ft binary.FuncType) interface{} {
	if len(ft.ResultTypes) == 0 {
		return nil
	}
	switch ft.ResultTypes[0] {
	case binary.ValTypeI32:
		return vm.popS32()
	case binary.ValTypeI64:
		return vm.popS64()
	case binary.ValTypeF32:
		return vm.popF32()
	case binary.ValTypeF64:
		return vm.popF64()
	default:
		panic("unreachable")
	}
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

func (vm *vm) CallFunc(name string, args ...interface{}) (interface{}, error) {
	fIdx, ok := vm.getFunc(name) // TODO
	if !ok {
		return nil, fmt.Errorf("function not found: " + name)
	}

	return vm.safeCallFunc(vm.funcs[fIdx], args)
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
