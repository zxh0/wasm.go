// +build jit

package jitllvm

import (
	"fmt"

	"github.com/tinygo-org/go-llvm"
	"github.com/zxh0/wasm.go/binary"
)

type compiler struct {
	module      binary.Module
	funcImports []binary.Import
	llvmMod     llvm.Module
	builder     llvm.Builder
	curFunc     llvm.Value
	funcs       []llvm.Value
	globals     []llvm.Value
	operands    []llvm.Value
	blockDepth  int
	reg         int
}

func newCompiler(module binary.Module) *compiler {
	return &compiler{
		module:  module,
		llvmMod: llvm.NewModule("jit"),
		builder: llvm.NewBuilder(),
	}
}

func Compile(module binary.Module) {
	newCompiler(module).compileModule()
}

func (c *compiler) nexRegName() string {
	c.reg++
	return fmt.Sprintf("v%d", c.reg-1)
}
func (c *compiler) pushOperand(val llvm.Value) {
	c.operands = append(c.operands, val)
}
func (c *compiler) popOperand() llvm.Value {
	val := c.operands[len(c.operands)-1]
	c.operands = c.operands[:len(c.operands)-1]
	return val
}
func (c *compiler) topOperand() llvm.Value {
	return c.operands[len(c.operands)-1]
}

func (c *compiler) compileModule() {
	c.globals = make([]llvm.Value, len(c.module.GlobalSec))
	for i, g := range c.module.GlobalSec {
		gt := valType2LLVM(g.Type.ValType)
		name := fmt.Sprintf("g%d", i)
		c.globals[i] = llvm.AddGlobal(c.llvmMod, gt, name)
		// TODO: init
	}
	for _, imp := range c.module.ImportSec {
		if imp.Desc.Tag == binary.ImportTagFunc {
			c.funcImports = append(c.funcImports, imp)
			name := fmt.Sprintf("f%d", len(c.funcs))
			ft := funcType2LLVM(c.module.TypeSec[imp.Desc.FuncType])
			f := llvm.AddFunction(c.llvmMod, name, ft)
			c.funcs = append(c.funcs, f)
		}
	}
	for _, ftIdx := range c.module.FuncSec {
		name := fmt.Sprintf("f%d", len(c.funcs))
		ft := funcType2LLVM(c.module.TypeSec[ftIdx])
		f := llvm.AddFunction(c.llvmMod, name, ft)
		c.funcs = append(c.funcs, f)
	}
	for i, code := range c.module.CodeSec {
		c.curFunc = c.funcs[len(c.funcImports)+i]
		c.compileCode(code)
	}
	fmt.Println(c.llvmMod.String())
}

func (c *compiler) compileCode(code binary.Code) {
	c.operands = c.operands[0:0]
	c.blockDepth = 0
	c.reg = 0

	entry := llvm.AddBasicBlock(c.curFunc, "entry")
	c.builder.SetInsertPoint(entry, entry.FirstInstruction())
	for i, param := range c.curFunc.Params() {
		param.SetName(fmt.Sprintf("p%d", i))
		p := c.builder.CreateAlloca(param.Type(), c.nexRegName())
		c.builder.CreateStore(param, p)
		c.pushOperand(p)
	}
	for _, locals := range code.Locals {
		for j := 0; j < int(locals.N); j++ {
			localType := valType2LLVM(locals.Type)
			p := c.builder.CreateAlloca(localType, c.nexRegName())
			//c.builder.CreateStore(llvm.ConstInt(localType, 0, false), v)
			c.pushOperand(p)
		}
	}

	bb, _ := c.compileBlock(0, 0, code.Expr)
	c.linkBBs(entry, bb)
}

func (c *compiler) compileBlock(depth, blockId int, instrs []binary.Instruction) (bb, bbEnd llvm.BasicBlock) {
	return c.compileInstrs(depth, blockId, 0, instrs)
}

func (c *compiler) compileInstrs(depth, blockId, pc int,
	instrs []binary.Instruction) (bb, bbEnd llvm.BasicBlock) {

	idx := findCtrlInstr(instrs)
	if idx < 0 {
		bb = c.emitBasicBlock(depth, blockId, pc, instrs)
		bbEnd = llvm.AddBasicBlock(c.curFunc,
			fmt.Sprintf("d%d_b%d_end", depth, blockId))
		c.linkBBs(bb, bbEnd)
		return
	} else if idx > 0 {
		var bbNext llvm.BasicBlock
		bb = c.emitBasicBlock(depth, blockId, pc, instrs[:idx])
		bbNext, bbEnd = c.compileInstrs(depth, blockId, pc+idx, instrs[idx:])
		c.linkBBs(bb, bbNext)
		return bb, bbEnd
	} else if instrs[0].Opcode == binary.Block ||
		instrs[0].Opcode == binary.Loop {
		bb, bbEnd = c.compileBlock(depth+1, pc, instrs[0].Args.(binary.BlockArgs).Instrs)
		if len(instrs) == 1 {
			return
		}
		bbNext, bbEnd2 := c.compileInstrs(depth, blockId, pc+1, instrs[1:])
		c.linkBBs(bbEnd, bbNext)
		return bb, bbEnd2
	} else if instrs[0].Opcode == binary.If {
		panic("TODO")
	} else if instrs[0].Opcode == binary.Br {
		panic("TODO")
	} else if instrs[0].Opcode == binary.BrIf {
		panic("TODO")
	} else if instrs[0].Opcode == binary.BrTable {
		panic("TODO")
	} else if instrs[0].Opcode == binary.Return {
		panic("TODO")
	} else {
		panic("TODO")
	}
}

func (c *compiler) linkBBs(bb1, bb2 llvm.BasicBlock) {
	c.builder.SetInsertPointAtEnd(bb1)
	c.builder.CreateBr(bb2)
	//c.builder.SetInsertPointAtEnd(bb2)
}

func (c *compiler) emitBasicBlock(depth, blockId, pc int, instrs []binary.Instruction) llvm.BasicBlock {
	bbName := fmt.Sprintf("d%d_b%d_pc%d", depth, blockId, pc)
	bb := llvm.AddBasicBlock(c.curFunc, bbName)
	c.builder.SetInsertPoint(bb, bb.FirstInstruction())
	for _, instr := range instrs {
		c.emitNonCtrlInstr(instr)
	}
	return bb
}

func (c *compiler) emitNonCtrlInstr(instr binary.Instruction) {
	//llvm.ExecutionEngine{}.AddGlobalMapping()
	switch instr.Opcode {
	case binary.Call:
		fnIdx := instr.Args.(uint32)
		ft := c.getFuncType(int(fnIdx))
		fn := c.funcs[fnIdx]
		args := c.operands[len(c.operands)-len(ft.ParamTypes):]
		rv := c.builder.CreateCall(fn, args, c.nexRegName())
		if len(ft.ResultTypes) > 0 {
			c.pushOperand(rv)
		}
	case binary.Drop:
		c.popOperand()
	case binary.LocalGet:
		p := c.operands[instr.Args.(uint32)]
		v := c.builder.CreateLoad(p, c.nexRegName())
		c.pushOperand(v)
	case binary.LocalSet:
		p := c.operands[instr.Args.(uint32)]
		c.builder.CreateStore(c.popOperand(), p)
	case binary.LocalTee:
		p := c.operands[instr.Args.(uint32)]
		c.builder.CreateStore(c.topOperand(), p)
	case binary.GlobalGet:
		p := c.globals[instr.Args.(uint32)]
		v := c.builder.CreateLoad(p, c.nexRegName())
		c.pushOperand(v)
	case binary.GlobalSet:
		p := c.globals[instr.Args.(uint32)]
		c.builder.CreateStore(c.popOperand(), p)
	case binary.I32Const:
		k := constI32(instr.Args.(int32))
		c.pushOperand(k)
	case binary.I64Const:
		k := constI64(instr.Args.(int64))
		c.pushOperand(k)
	case binary.I32Add:
		rhs, lhs := c.popOperand(), c.popOperand()
		v := c.builder.CreateAdd(rhs, lhs, c.nexRegName())
		c.pushOperand(v)
	default:
		panic("TODO: " + instr.GetOpname())
	}
}

func (c *compiler) getFuncType(idx int) binary.FuncType {
	if idx < len(c.funcImports) {
		ftIdx := c.funcImports[idx].Desc.FuncType
		return c.module.TypeSec[ftIdx]
	} else {
		ftIdx := c.module.FuncSec[idx-len(c.funcImports)]
		return c.module.TypeSec[ftIdx]
	}
}
