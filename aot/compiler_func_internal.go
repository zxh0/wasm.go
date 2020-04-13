package aot

import (
	"fmt"
	"math"
	"strings"

	"github.com/zxh0/wasm.go/binary"
)

type internalFuncCompiler struct {
	funcCompiler
	moduleInfo moduleInfo
	stackPtr   int
	stackMax   int
	tmpIdx     int
	blocks     []blockInfo
}

type blockInfo struct {
	isLoop    bool
	hasResult bool
	stackPtr  int
}

func newInternalFuncCompiler(moduleInfo moduleInfo) *internalFuncCompiler {
	return &internalFuncCompiler{
		funcCompiler: newFuncCompiler(),
		moduleInfo:   moduleInfo,
	}
}

func (c *internalFuncCompiler) printIndentsPlus(n int) {
	for i := len(c.blocks) + n; i > 0; i-- {
		c.print("\t")
	}
}
func (c *internalFuncCompiler) printIndents() {
	for i := len(c.blocks); i > 0; i-- {
		c.print("\t")
	}
}

func (c *internalFuncCompiler) stackPush() int {
	c.stackPtr++
	if c.stackMax < c.stackPtr {
		c.stackMax = c.stackPtr
	}
	return c.stackPtr - 1
}
func (c *internalFuncCompiler) stackPop() int {
	c.stackPtr--
	return c.stackPtr + 1
}
func (c *internalFuncCompiler) stackTop() int {
	return c.stackPtr - 1
}

func (c *internalFuncCompiler) enterBlock(isLoop, hasResult bool) {
	c.blocks = append(c.blocks, blockInfo{
		isLoop:    isLoop,
		hasResult: hasResult,
		stackPtr:  c.stackPtr,
	})
}
func (c *internalFuncCompiler) exitBlock() {
	c.blocks = c.blocks[:len(c.blocks)-1]
}
func (c *internalFuncCompiler) blockDepth() int {
	return len(c.blocks)
}

func (c *internalFuncCompiler) compile(idx int,
	ft binary.FuncType, code binary.Code) string {

	paramCount := len(ft.ParamTypes)
	resultCount := len(ft.ResultTypes)
	localCount := int(code.GetLocalCount())

	c.stackPtr = paramCount + localCount
	c.stackMax = c.stackPtr

	c.printf("func (m *aotModule) f%d(", idx)
	c.genParams(paramCount)
	c.print(")")
	c.genResults(resultCount)
	c.print(" {\n")
	c.println("	// var ... uint64")
	c.genFuncBody(code, resultCount)
	c.println("}")

	s := c.sb.String()
	if c.stackMax > paramCount {
		s = strings.ReplaceAll(s, "// var ... uint64",
			genLocals(paramCount, localCount, c.stackMax))
	}
	if cheat := genUnusedSlotsCheat(s, resultCount, c.stackMax); cheat != "" {
		s = s[:len(s)-2] + cheat + "}\n"
	}
	return s
}

func genLocals(paramCount, localCount, stackMax int) string {
	p := newPrinter()
	p.print("var ")
	for i := paramCount; i < stackMax; i++ {
		p.printIf(i > paramCount, ", ", "")
		p.printf("s%d", i)
	}
	p.printf(" uint64 // locals: %d", localCount)
	return p.sb.String()
}
func genUnusedSlotsCheat(s string, resultCount, stackMax int) string {
	s = s[strings.Index(s, "// locals:")+10:]

	var unusedSlots []int
	for i := 0; i < stackMax; i++ {
		if !isSlotUsed(s, i) {
			unusedSlots = append(unusedSlots, i)
		}
	}
	if unusedSlots == nil {
		return ""
	}

	p := newPrinter()
	p.print("\treturn")
	p.printIf(resultCount > 0, " 0; ", "; ")
	p.print("print(")
	for _, slot := range unusedSlots {
		p.printf("s%d, ", slot)
	}
	p.println(") // 'xxx declared and not used' cheat")
	return p.sb.String()
}
func isSlotUsed(s string, slotIdx int) bool {
	slotName := fmt.Sprintf("s%d", slotIdx)
	assign := slotName + " = "
	for len(s) > 0 {
		idx := strings.Index(s, slotName)
		if idx < 0 {
			return false
		}
		s = s[idx:]
		if !strings.HasPrefix(s, assign) {
			return true
		}
		s = s[len(assign):]
	}
	return false
}

func (c *internalFuncCompiler) genFuncBody(code binary.Code, resultCount int) {
	expr := analyzeBr(code)
	c.emitBlock(expr, false, resultCount > 0)
	if resultCount > 0 {
		c.printf("\treturn s%d\n", c.stackPtr-1)
	}
}

func (c *internalFuncCompiler) emitInstr(instr binary.Instruction) {
	switch instr.Opcode {
	case binary.Block, binary.Loop, binary.If:
	case binary.BrTable:
	case 0xFF:
	default:
		c.printIndents()
	}

	opname := instr.GetOpname()
	switch instr.Opcode {
	case binary.Unreachable:
		c.printf(`panic("unreachable") // %s\n`, opname) // TODO
	case binary.Nop:
		c.printf("// %s\n", opname)
	case binary.Block:
		blockArgs := instr.Args.(binary.BlockArgs)
		rt := c.moduleInfo.module.GetResultTypes(blockArgs.RT)
		c.emitBlock(blockArgs.Instrs, false, len(rt) > 0)
	case binary.Loop:
		blockArgs := instr.Args.(binary.BlockArgs)
		rt := c.moduleInfo.module.GetResultTypes(blockArgs.RT)
		c.emitBlock(blockArgs.Instrs, true, len(rt) > 0)
	case binary.If:
		c.emitIf(instr.Args.(binary.IfArgs))
	case binary.Br:
		c.emitBr(instr.Args.(uint32))
	case binary.BrIf:
		c.emitBrIf(instr.Args.(uint32))
	case binary.BrTable:
		c.emitBrTable(instr.Args.(binary.BrTableArgs))
	case binary.Return:
		c.emitReturn()
	case binary.Call:
		c.emitCall(int(instr.Args.(uint32)))
	case binary.CallIndirect:
		c.emitCallIndirect(int(instr.Args.(uint32)))
	case binary.Drop:
		c.printf("// %s\n", opname)
		c.stackPop()
	case binary.Select:
		c.printf("if s%d == 0 { s%d = s%d } // %s\n",
			c.stackPtr-1, c.stackPtr-3, c.stackPtr-2, opname)
		c.stackPtr -= 2
	case binary.LocalGet:
		c.printf("s%d = s%d // %s %d\n",
			c.stackPush(), instr.Args, opname, instr.Args)
	case binary.LocalSet:
		c.printf("s%d = s%d // %s %d\n",
			instr.Args, c.stackPtr-1, opname, instr.Args)
		c.stackPtr--
	case binary.LocalTee:
		c.printf("s%d = s%d // %s %d\n",
			instr.Args, c.stackPtr-1, opname, instr.Args)
	case binary.GlobalGet:
		c.printf("s%d = m.globals[%d].Get() // %s %d\n",
			c.stackPush(), instr.Args, opname, instr.Args)
	case binary.GlobalSet:
		c.printf("m.globals[%d].Set(s%d) // %s %d\n",
			instr.Args, c.stackPtr-1, opname, instr.Args)
		c.stackPtr--
	case binary.I32Load, binary.F32Load:
		c.emitLoad(instr, "s%d = uint64(m.readU32(s%d + %d)) // %s\n")
	case binary.I64Load, binary.F64Load:
		c.emitLoad(instr, "s%d = m.readU64(s%d + %d) // %s\n")
	case binary.I32Load8S:
		c.emitLoad(instr, "s%d = uint64(int8(m.readU8(s%d + %d))) // %s\n")
	case binary.I32Load8U:
		c.emitLoad(instr, "s%d = uint64(m.readU8(s%d + %d)) // %s\n")
	case binary.I32Load16S:
		c.emitLoad(instr, "s%d = uint64(int16(m.readU16(s%d + %d))) // %s\n")
	case binary.I32Load16U:
		c.emitLoad(instr, "s%d = uint64(m.readU16(s%d + %d)) // %s\n")
	case binary.I64Load8S:
		c.emitLoad(instr, "s%d = uint64(int8(m.readU8(s%d + %d))) // %s\n")
	case binary.I64Load8U:
		c.emitLoad(instr, "s%d = uint64(m.readU8(s%d + %d)) // %s\n")
	case binary.I64Load16S:
		c.emitLoad(instr, "s%d = uint64(int16(m.readU16(s%d + %d))) // %s\n")
	case binary.I64Load16U:
		c.emitLoad(instr, "s%d = uint64(m.readU16(s%d + %d)) // %s\n")
	case binary.I64Load32S:
		c.emitLoad(instr, "s%d = uint64(int32(m.readU32(s%d + %d))) // %s\n")
	case binary.I64Load32U:
		c.emitLoad(instr, "s%d = uint64(m.readU32(s%d + %d)) // %s\n")
	case binary.I32Store, binary.F32Store:
		c.emitStore(instr, "m.writeU32(s%d + %d, uint32(s%d)) // %s\n")
	case binary.I64Store, binary.F64Store:
		c.emitStore(instr, "m.writeU64(s%d + %d, s%d) // %s\n")
	case binary.I32Store8, binary.I64Store8:
		c.emitStore(instr, "m.writeU8(s%d + %d, byte(s%d)) // %s\n")
	case binary.I32Store16, binary.I64Store16:
		c.emitStore(instr, "m.writeU16(s%d + %d, uint16(s%d)) // %s\n")
	case binary.I64Store32:
		c.emitStore(instr, "m.writeU32(s%d + %d, uint32(s%d)) // %s\n")
	case binary.MemorySize:
		c.emitMemSize(opname)
	case binary.MemoryGrow:
		c.emitMemGrow(opname)
	case binary.I32Const:
		c.emitConst(uint64(uint32(instr.Args.(int32))), opname, instr.Args)
	case binary.I64Const:
		c.emitConst(uint64(instr.Args.(int64)), opname, instr.Args)
	case binary.F32Const:
		c.emitConst(uint64(math.Float32bits(instr.Args.(float32))), opname, instr.Args)
	case binary.F64Const:
		c.emitConst(math.Float64bits(instr.Args.(float64)), opname, instr.Args)
	case binary.I32Eqz:
		c.printf("s%d = b2i(uint32(s%d) == 0) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32Eq:
		c.emitI32BinCmpU("==", opname)
	case binary.I32Ne:
		c.emitI32BinCmpU("!=", opname)
	case binary.I32LtS:
		c.emitI32BinCmpS("<", opname)
	case binary.I32LtU:
		c.emitI32BinCmpU("<", opname)
	case binary.I32GtS:
		c.emitI32BinCmpS(">", opname)
	case binary.I32GtU:
		c.emitI32BinCmpU(">", opname)
	case binary.I32LeS:
		c.emitI32BinCmpS("<=", opname)
	case binary.I32LeU:
		c.emitI32BinCmpU("<=", opname)
	case binary.I32GeS:
		c.emitI32BinCmpS(">=", opname)
	case binary.I32GeU:
		c.emitI32BinCmpU(">=", opname)
	case binary.I64Eqz:
		c.printf("s%d = b2i(s%d == 0) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64Eq:
		c.emitI64BinCmpU("==", opname)
	case binary.I64Ne:
		c.emitI64BinCmpU("!=", opname)
	case binary.I64LtS:
		c.emitI64BinCmpS("<", opname)
	case binary.I64LtU:
		c.emitI64BinCmpU("<", opname)
	case binary.I64GtS:
		c.emitI64BinCmpS(">", opname)
	case binary.I64GtU:
		c.emitI64BinCmpU(">", opname)
	case binary.I64LeS:
		c.emitI64BinCmpS("<=", opname)
	case binary.I64LeU:
		c.emitI64BinCmpU("<=", opname)
	case binary.I64GeS:
		c.emitI64BinCmpS(">=", opname)
	case binary.I64GeU:
		c.emitI32BinCmpU(">=", opname)
	case binary.F32Eq:
		c.emitF32BinCmp("==", opname)
	case binary.F32Ne:
		c.emitF32BinCmp("!=", opname)
	case binary.F32Lt:
		c.emitF32BinCmp("<", opname)
	case binary.F32Gt:
		c.emitF32BinCmp(">", opname)
	case binary.F32Le:
		c.emitF32BinCmp("<=", opname)
	case binary.F32Ge:
		c.emitF32BinCmp(">=", opname)
	case binary.F64Eq:
		c.emitF64BinCmp("==", opname)
	case binary.F64Ne:
		c.emitF64BinCmp("!=", opname)
	case binary.F64Lt:
		c.emitF64BinCmp("<", opname)
	case binary.F64Gt:
		c.emitF64BinCmp(">", opname)
	case binary.F64Le:
		c.emitF64BinCmp("<=", opname)
	case binary.F64Ge:
		c.emitF64BinCmp(">=", opname)
	case binary.I32Clz:
		c.printf("s%d = uint64(bits.LeadingZeros32(uint32(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32Ctz:
		c.printf("s%d = uint64(bits.TrailingZeros32(uint32(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32PopCnt:
		c.printf("s%d = uint64(bits.OnesCount32(uint32(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32Add:
		c.emitI32BinArithU("+", opname)
	case binary.I32Sub:
		c.emitI32BinArithU("-", opname)
	case binary.I32Mul:
		c.emitI32BinArithU("*", opname)
	case binary.I32DivS:
		c.emitI32BinArithS("/", opname)
	case binary.I32DivU:
		c.emitI32BinArithU("/", opname)
	case binary.I32RemS:
		c.emitI32BinArithS("%", opname)
	case binary.I32RemU:
		c.emitI32BinArithU("/", opname)
	case binary.I32And:
		c.emitI32BinArithU("&", opname)
	case binary.I32Or:
		c.emitI32BinArithU("|", opname)
	case binary.I32Xor:
		c.emitI32BinArithU("^", opname)
	case binary.I32Shl:
		c.printf("s%d = uint64(uint32(s%d) << (uint32(s%d) %% 32)) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I32ShrS:
		c.printf("s%d = uint64(int32(s%d) >> (uint32(s%d) %% 32)) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I32ShrU:
		c.printf("s%d = uint64(uint32(s%d) >> (uint32(s%d) %% 32)) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I32Rotl:
		c.printf("s%d = uint64(bits.RotateLeft32(uint32(s%d), int(uint32(s%d)))) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I32Rotr:
		c.printf("s%d = uint64(bits.RotateLeft32(uint32(s%d), -int(uint32(s%d)))) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I64Clz:
		c.printf("s%d = uint64(bits.LeadingZeros64(s%d)) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64Ctz:
		c.printf("s%d = uint64(bits.TrailingZeros64(s%d)) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64PopCnt:
		c.printf("s%d = uint64(bits.OnesCount64(s%d)) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64Add:
		c.emitI64BinArithU("+", opname)
	case binary.I64Sub:
		c.emitI64BinArithU("-", opname)
	case binary.I64Mul:
		c.emitI64BinArithU("*", opname)
	case binary.I64DivS:
		c.emitI64BinArithS("/", opname)
	case binary.I64DivU:
		c.emitI64BinArithU("/", opname)
	case binary.I64RemS:
		c.emitI64BinArithS("%", opname)
	case binary.I64RemU:
		c.emitI64BinArithU("/", opname)
	case binary.I64And:
		c.emitI64BinArithU("&", opname)
	case binary.I64Or:
		c.emitI64BinArithU("|", opname)
	case binary.I64Xor:
		c.emitI64BinArithU("^", opname)
	case binary.I64Shl:
		c.printf("s%d = s%d << (s%d %% 64) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I64ShrS:
		c.printf("s%d = uint64(int64(s%d) >> (s%d %% 64)) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I64ShrU:
		c.printf("s%d = s%d >> (s%d %% 64) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I64Rotl:
		c.printf("s%d = bits.RotateLeft64(s%d, int(s%d)) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I64Rotr:
		c.printf("s%d = bits.RotateLeft64(s%d, int(s%d)) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.F32Abs:
		c.emitF32UnFC("math.Abs", opname)
	case binary.F32Neg:
		c.emitF32UnFC("-", opname)
	case binary.F32Ceil:
		c.emitF32UnFC("math.Ceil", opname)
	case binary.F32Floor:
		c.emitF32UnFC("math.Floor", opname)
	case binary.F32Trunc:
		c.emitF32UnFC("math.Trunc", opname)
	case binary.F32Nearest:
		c.emitF32UnFC("math.RoundToEven", opname)
	case binary.F32Sqrt:
		c.emitF32UnFC("math.Sqrt", opname)
	case binary.F32Add:
		c.emitF32BinArith("+", opname)
	case binary.F32Sub:
		c.emitF32BinArith("-", opname)
	case binary.F32Mul:
		c.emitF32BinArith("*", opname)
	case binary.F32Div:
		c.emitF32BinArith("/", opname)
	case binary.F32Min:
		c.emitF32BinFC("math.Min", opname)
	case binary.F32Max:
		c.emitF32BinFC("math.Max", opname)
	case binary.F32CopySign:
		c.emitF32BinFC("math.Copysign", opname)
	case binary.F64Abs:
		c.emitF64UnFC("math.Abs", opname)
	case binary.F64Neg:
		c.emitF64UnFC("-", opname)
	case binary.F64Ceil:
		c.emitF64UnFC("math.Ceil", opname)
	case binary.F64Floor:
		c.emitF64UnFC("math.Floor", opname)
	case binary.F64Trunc:
		c.emitF64UnFC("math.Trunc", opname)
	case binary.F64Nearest:
		c.emitF64UnFC("math.RoundToEven", opname)
	case binary.F64Sqrt:
		c.emitF64UnFC("math.Sqrt", opname)
	case binary.F64Add:
		c.emitF64BinArith("+", opname)
	case binary.F64Sub:
		c.emitF64BinArith("-", opname)
	case binary.F64Mul:
		c.emitF64BinArith("*", opname)
	case binary.F64Div:
		c.emitF64BinArith("/", opname)
	case binary.F64Min:
		c.emitF64BinFC("math.Min", opname)
	case binary.F64Max:
		c.emitF64BinFC("math.Max", opname)
	case binary.F64CopySign:
		c.emitF64BinFC("math.Copysign", opname)
	case binary.I32WrapI64:
		c.printf("s%d = uint64(uint32(s%d)) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32TruncF32S:
		c.printf("s%d = uint64(uint32(int32(math.Trunc(float64(_f32(s%d)))))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32TruncF32U:
		c.printf("s%d = uint64(uint32(math.Trunc(float64(_f32(s%d))))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32TruncF64S:
		c.printf("s%d = uint64(uint32(int32(math.Trunc(_f64(s%d))))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32TruncF64U:
		c.printf("s%d = uint64(uint32(math.Trunc(_f64(s%d)))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64ExtendI32S:
		c.printf("s%d = uint64(int64(int32(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64ExtendI32U:
		c.printf("s%d = uint64(uint32(s%d)) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64TruncF32S:
		c.printf("s%d = uint64(int64(math.Trunc(float64(_f32(s%d))))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64TruncF32U:
		c.printf("s%d = uint64(math.Trunc(float64(_f32(s%d)))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64TruncF64S:
		c.printf("s%d = uint64(int64(math.Trunc(_f64(s%d)))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64TruncF64U:
		c.printf("s%d = uint64(math.Trunc(_f64(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F32ConvertI32S:
		c.printf("s%d = _u32(float32(int32(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F32ConvertI32U:
		c.printf("s%d = _u32(float32(uint32(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F32ConvertI64S:
		c.printf("s%d = _u32(float32(int64(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F32ConvertI64U:
		c.printf("s%d = _u32(float32(s%d)) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F32DemoteF64:
		c.printf("s%d = _u32(float32(_f64(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F64ConvertI32S:
		c.printf("s%d = _u64(float64(int32(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F64ConvertI32U:
		c.printf("s%d = _u64(float64(uint32(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F64ConvertI64S:
		c.printf("s%d = _u64(float64(int64(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F64ConvertI64U:
		c.printf("s%d = _u64(float64(s%d)) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F64PromoteF32:
		c.printf("s%d = _u64(float64(_f32(s%d))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32ReinterpretF32:
		c.printf("// %s\n", opname) // TODO
	case binary.I64ReinterpretF64:
		c.printf("// %s\n", opname) // TODO
	case binary.F32ReinterpretI32:
		c.printf("// %s\n", opname) // TODO
	case binary.F64ReinterpretI64:
		c.printf("// %s\n", opname) // TODO
	case 0xFF:
	default:
		c.printf("// 0x%X ???\n", instr.Opcode)
	}
}

/*
l0: for {
	... // break
	break
}
*/
func (c *internalFuncCompiler) emitBlock(expr []binary.Instruction, isLoop, hasResult bool) {
	c.printIndents()
	if isBrTarget(expr) {
		c.printf("_l%d: for {\n", c.blockDepth())
	} else {
		c.printf("{ // _l%d\n", c.blockDepth())
	}
	c.enterBlock(isLoop, hasResult)
	for _, instr := range expr {
		c.emitInstr(instr)
	}
	c.exitBlock()
	if isBrTarget(expr) {
		c.printIndentsPlus(1)
		c.printf("break _l%d\n", c.blockDepth())
		c.printIndents()
		c.printf("} // end of _l%d\n", c.blockDepth())
	} else {
		c.printIndents()
		c.printf("} // end of _l%d\n", c.blockDepth())
	}
}

/*
l0: for {
	... // continue
	break
}
*/
func (c *internalFuncCompiler) emitLoop(blockArgs binary.BlockArgs) {
	rt := c.moduleInfo.module.GetResultTypes(blockArgs.RT)
	c.emitBlock(blockArgs.Instrs, true, len(rt) > 0)
}

/*
l0: for {
	if <cond> {
		...
	} else {
		...
	}
}
*/
func (c *internalFuncCompiler) emitIf(ifArgs binary.IfArgs) {
	if isBrTarget(ifArgs.Instrs1) {
		c.printIndents()
		c.printf("_l%d: for {\n", c.blockDepth()-1)
	}
	rt := c.moduleInfo.module.GetResultTypes(ifArgs.RT)
	c.enterBlock(false, len(rt) > 0)

	c.printIndentsPlus(-1)
	c.printf("if s%d > 0 { // if@%d\n", c.stackPtr-1, len(c.blocks)-1)
	c.stackPop()
	stackPtr := c.stackPtr
	for _, instr := range ifArgs.Instrs1 {
		c.emitInstr(instr)
	}
	c.stackPtr = stackPtr
	if len(ifArgs.Instrs2) > 0 {
		c.printIndentsPlus(-1)
		c.println("} else {")
	}
	for _, instr := range ifArgs.Instrs2 {
		c.emitInstr(instr)
	}
	c.printIndentsPlus(-1)
	c.printf("} // end if@%d\n", len(c.blocks)-1)

	c.exitBlock()
	if isBrTarget(ifArgs.Instrs1) {
		c.printIndents()
		c.printf("break } // end of _l%d\n", c.blockDepth()-1)
	}
}

func (c *internalFuncCompiler) emitBr(labelIdx uint32) {
	n := len(c.blocks) - int(labelIdx) - 1
	c.printIf(c.blocks[n].isLoop, "continue ", "break ")
	c.printf("_l%d // br %d\n", n, labelIdx)
}
func (c *internalFuncCompiler) emitBrIf(labelIdx uint32) {
	n := len(c.blocks) - int(labelIdx) - 1
	br := "break"
	if c.blocks[n].isLoop {
		br = "continue"
	}
	c.printf("if s%d != 0 { %s _l%d } // br_if %d\n",
		c.stackPtr-1, br, n, labelIdx)
	c.stackPop()
}
func (c *internalFuncCompiler) emitBrTable(btArgs binary.BrTableArgs) {
	c.printIndents()
	c.printf("// br_table %v %d\n", btArgs.Labels, btArgs.Default)
	for i, label := range btArgs.Labels {
		c.printIndents()
		c.printIf(i > 0, "} else ", "")
		c.printf("if s%d == %d {\n", c.stackPtr-1, i)
		c.printIndentsPlus(1)
		n := len(c.blocks) - int(label) - 1
		c.printIf(c.blocks[n].isLoop, "continue ", "break ")
		c.printf("_l%d //\n", n)
	}
	c.printIndents()
	c.println("}")
}
func (c *internalFuncCompiler) emitReturn() {
	c.printf("return s%d // return\n", c.stackPtr-1)
}

func (c *internalFuncCompiler) emitCall(funcIdx int) {
	ft := c.moduleInfo.getFuncType(funcIdx)
	c.stackPtr -= len(ft.ParamTypes)
	if len(ft.ResultTypes) > 0 {
		c.printf("s%d = ", c.stackPtr)
	}
	c.printf("m.f%d(", funcIdx)
	for i := range ft.ParamTypes {
		c.printIf(i > 0, ", ", "")
		c.printf("s%d", c.stackPtr+i)
	}
	if len(ft.ResultTypes) > 0 {
		c.stackPtr++
	}
	c.printf(") // call func#%d\n", funcIdx)
}
func (c *internalFuncCompiler) emitCallIndirect(typeIdx int) {
	elemIdx := c.stackPtr - 1
	c.stackPop()

	ft := c.moduleInfo.module.TypeSec[typeIdx]
	c.stackPtr -= len(ft.ParamTypes)
	if len(ft.ResultTypes) > 0 {
		c.printf("t%d, _ := ", c.tmpIdx)
		c.tmpIdx++
	}

	c.printf("m.table.GetElem(uint32(s%d)).Call(", elemIdx)
	for i := range ft.ParamTypes {
		c.printIf(i > 0, ", ", "")
		c.printf("s%d", c.stackPtr+i) // TODO
	}
	c.printf(") // call_indirect type#%d\n", typeIdx)

	if len(ft.ResultTypes) > 0 {
		c.printIndents()
		switch ft.ResultTypes[0] {
		case binary.ValTypeI32:
			c.printf("s%d = uint64(t%d.(int32))\n", c.stackPtr, c.tmpIdx-1)
		case binary.ValTypeI64:
			c.printf("s%d = uint64(t%d.(int64))\n", c.stackPtr, c.tmpIdx-1)
		case binary.ValTypeF32:
			c.printf("s%d = _u32(t%d.(float32))\n", c.stackPtr, c.tmpIdx-1)
		case binary.ValTypeF64:
			c.printf("s%d = _u64(t%d.(float64))\n", c.stackPtr, c.tmpIdx-1)
		}
		c.stackPtr++
	}
}

func (c *internalFuncCompiler) emitLoad(instr binary.Instruction, tmpl string) {
	// s%d = m.readU32(s%d + %d) // %s\n
	c.printf(tmpl, c.stackPtr-1, c.stackPtr-1, instr.Args.(binary.MemArg).Offset, instr.GetOpname())
}
func (c *internalFuncCompiler) emitStore(instr binary.Instruction, tmpl string) {
	// m.writeU32(s%d + %d, uint32(s%d)) // %s\n
	c.printf(tmpl, c.stackPtr-2, instr.Args.(binary.MemArg).Offset, c.stackPtr-1, instr.GetOpname())
	c.stackPtr -= 2
}
func (c *internalFuncCompiler) emitMemSize(opname string) {
	c.printf("s%d = uint64(m.memory.Size()) // %s\n",
		c.stackPush(), opname)
}
func (c *internalFuncCompiler) emitMemGrow(opname string) {
	c.printf("s%d = uint64(m.memory.Grow(uint32(s%d))) // %s\n",
		c.stackPtr-1, c.stackPtr-1, opname)
}

func (c *internalFuncCompiler) emitConst(val uint64, opname string, arg interface{}) {
	c.printf("s%d = 0x%x // %s %v\n",
		c.stackPush(), val, opname, arg)
}

func (c *internalFuncCompiler) emitI32BinCmpU(operator, opname string) {
	c.printf("s%d = b2i(uint32(s%d) %s uint32(s%d)) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *internalFuncCompiler) emitI32BinCmpS(operator, opname string) {
	c.printf("s%d = b2i(int32(s%d) %s int32(s%d)) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *internalFuncCompiler) emitI32BinArithU(operator, opname string) {
	c.printf("s%d = uint64(uint32(s%d) %s uint32(s%d)) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *internalFuncCompiler) emitI32BinArithS(operator, opname string) {
	c.printf("s%d = uint64(int32(s%d) %s int32(s%d)) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}

func (c *internalFuncCompiler) emitI64BinCmpU(operator, opname string) {
	c.printf("s%d = b2i(s%d %s s%d) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *internalFuncCompiler) emitI64BinCmpS(operator, opname string) {
	c.printf("s%d = b2i(int64(s%d) %s int64(s%d)) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *internalFuncCompiler) emitI64BinArithU(operator, opname string) {
	c.printf("s%d = s%d %s s%d // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *internalFuncCompiler) emitI64BinArithS(operator, opname string) {
	c.printf("s%d = uint64(int64(s%d) %s int64(s%d)) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}

func (c *internalFuncCompiler) emitF32BinCmp(operator, opname string) {
	c.printf("s%d = b2i(_f32(s%d) %s _f32(s%d)) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *internalFuncCompiler) emitF32BinArith(operator, opname string) {
	c.printf("s%d = _u32(_f32(s%d) %s _f32(s%d)) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *internalFuncCompiler) emitF32UnFC(funcName, opname string) {
	c.printf("s%d = _u32(float32(%s(float64(_f32(s%d))))) // %s\n",
		c.stackPtr-1, funcName, c.stackPtr-1, opname)
}
func (c *internalFuncCompiler) emitF32BinFC(funcName, opname string) {
	c.printf("s%d = _u32(float32(%s(float64(_f32(s%d)), float64(_f32(s%d))))) // %s\n",
		c.stackPtr-2, funcName, c.stackPtr-2, c.stackPtr-1, opname)
	c.stackPop()
}

func (c *internalFuncCompiler) emitF64BinCmp(operator, opname string) {
	c.printf("s%d = b2i(_f64(s%d) %s _f64(s%d)) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *internalFuncCompiler) emitF64BinArith(operator, opname string) {
	c.printf("s%d = _u64(_f64(s%d) %s _f64(s%d)) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *internalFuncCompiler) emitF64UnFC(funcName, opname string) {
	c.printf("s%d = _u64(%s(_f64(s%d))) // %s\n",
		c.stackPtr-1, funcName, c.stackPtr-1, opname)
}
func (c *internalFuncCompiler) emitF64BinFC(funcName, opname string) {
	c.printf("s%d = _u64(%s(_f64(s%d), _f64(s%d))) // %s\n",
		c.stackPtr-2, funcName, c.stackPtr-2, c.stackPtr-1, opname)
	c.stackPop()
}
