package aot

import (
	"fmt"
	"math"
	"strings"

	"github.com/zxh0/wasm.go/binary"
)

type funcCompiler struct {
	printer
	module   binary.Module
	stackPtr int
	stackMax int
	blocks   []blockInfo
}

type blockInfo struct {
	isLoop    bool
	hasResult bool
	stackPtr  int
}

func newFuncCompiler(module binary.Module) *funcCompiler {
	return &funcCompiler{
		printer: printer{sb: &strings.Builder{}},
		module:  module,
	}
}

func (c *funcCompiler) printIndents() {
	for i := len(c.blocks); i > 0; i-- {
		c.sb.WriteByte('\t')
	}
}

func (c *funcCompiler) stackPush() int {
	c.stackPtr++
	if c.stackMax < c.stackPtr {
		c.stackMax = c.stackPtr
	}
	return c.stackPtr - 1
}
func (c *funcCompiler) stackPop() int {
	c.stackPtr--
	return c.stackPtr + 1
}

func (c *funcCompiler) enterBlock(isLoop, hasResult bool) {
	c.blocks = append(c.blocks, blockInfo{
		isLoop:    isLoop,
		hasResult: hasResult,
		stackPtr:  c.stackPtr,
	})
}
func (c *funcCompiler) exitBlock() {
	c.blocks = c.blocks[len(c.blocks)-1:]
}
func (c *funcCompiler) blockDepth() int {
	return len(c.blocks)
}

func (c *funcCompiler) compile(idx int,
	ft binary.FuncType, code binary.Code) string {

	paramCount := len(ft.ParamTypes)
	resultCount := len(ft.ResultTypes)
	localCount := code.GetLocalCount()

	c.stackPtr = paramCount + localCount
	c.stackMax = c.stackPtr

	c.printf("func (m *aotModule) f%d(", idx)
	c.genParams(paramCount)
	c.print(")")
	c.genResults(resultCount)
	c.print(" {\n")
	c.genLocals(paramCount)
	c.genFuncBody(code, resultCount)
	c.println("}")

	stackMax := fmt.Sprintf("%d", c.stackMax)
	return strings.ReplaceAll(c.sb.String(), "$stackMax", stackMax)
}

func (c *funcCompiler) genParams(paramCount int) {
	for i := 0; i < paramCount; i++ {
		c.printf("p%d", i)
		if i < paramCount-1 {
			c.print(", ")
		} else {
			c.print(" uint64")
		}
	}
}
func (c *funcCompiler) genResults(resultCount int) {
	if resultCount == 1 {
		c.print(" uint64")
	}
}
func (c *funcCompiler) genLocals(paramCount int) {
	c.print("\tstack := [$stackMax]uint64{")
	for i := 0; i < paramCount; i++ {
		c.printf("p%d, ", i)
	}
	c.print("}\n")
}
func (c *funcCompiler) genFuncBody(code binary.Code, resultCount int) {
	c.emitBlock(code.Expr, false, resultCount > 0)
	if resultCount > 0 {
		c.printf("\treturn stack[%d]\n", c.stackPtr-1)
	}
}

func (c *funcCompiler) emitInstr(instr binary.Instruction) {
	opname := instr.String()
	c.printIndents()
	switch instr.Opcode {
	case binary.Unreachable:
		c.printf(`panic("TODO") // %s\n`, opname) // TODO
	case binary.Nop:
		c.printf("// %s\n", opname)
	case binary.Block:
		blockArgs := instr.Args.(binary.BlockArgs)
		c.emitBlock(blockArgs.Instrs, false, len(blockArgs.RT) > 0)
	case binary.Loop:
		c.emitLoop()
	case binary.If:
		c.emitIf()
	case binary.Br:
		c.emitBr(instr.Args.(uint32))
	case binary.BrIf:
		c.emitBrIf(instr.Args.(uint32))
	case binary.BrTable:
		c.emitBrTable()
	case binary.Return:
		c.emitReturn()
	case binary.Call:
		c.emitCall(instr.Args.(uint32), opname)
	case binary.CallIndirect:
		c.emitCallIndirect()
	case binary.Drop:
		c.printf("// %s\n", opname)
		c.stackPop()
	case binary.Select:
		c.printf("if stack[%d] == 0 { stack[%d] = stack[%d] } // %s\n",
			c.stackPtr-1, c.stackPtr-3, c.stackPtr-2, opname)
		c.stackPtr -= 2
	case binary.LocalGet:
		c.printf("stack[%d] = stack[%d] // %s\n",
			c.stackPush(), instr.Args.(uint32), opname)
	case binary.LocalSet:
		c.printf("stack[%d] = stack[%d] // %s\n",
			instr.Args.(uint32), c.stackPop(), opname)
	case binary.LocalTee:
		c.printf("stack[%d] = stack[%d] // %s\n",
			instr.Args.(uint32), c.stackPtr-1, opname)
	case binary.GlobalGet:
		c.printf("stack[%d] = m.globals[%d] // %s\n",
			c.stackPush(), instr.Args.(uint32), opname)
	case binary.GlobalSet:
		c.printf("m.globals[%d] = stack[%d] // %s\n",
			instr.Args.(uint32), c.stackPop(), opname)
	case binary.I32Load, binary.F32Load:
		c.emitLoad(instr, opname, "stack[%d] = binary.LittleEndian.Uint32(m.memory[stack[%d] + %d:]) // %s\n")
	case binary.I64Load, binary.F64Load:
		c.emitLoad(instr, opname, "stack[%d] = binary.LittleEndian.Uint64(m.memory[stack[%d] + %d:]) // %s\n")
	case binary.I32Load8S:
		c.emitLoad(instr, opname, "stack[%d] = uint32(int32(int8(m.memory[stack[%d] + %d:]))) // %s\n")
	case binary.I32Load8U:
		c.emitLoad(instr, opname, "stack[%d] = uint32(m.memory[stack[%d] + %d:]) // %s\n")
	case binary.I32Load16S:
		c.emitLoad(instr, opname, "stack[%d] = uint32(int32(int16(binary.LittleEndian.Uint16(m.memory[stack[%d] + %d:])))) // %s\n")
	case binary.I32Load16U:
		c.emitLoad(instr, opname, "stack[%d] = uint32(binary.LittleEndian.Uint16(m.memory[stack[%d] + %d:])) // %s\n")
	case binary.I64Load8S:
		c.emitLoad(instr, opname, "stack[%d] = uint64(int64(int8(m.memory[stack[%d] + %d:]))) // %s\n")
	case binary.I64Load8U:
		c.emitLoad(instr, opname, "stack[%d] = uint64(m.memory[stack[%d] + %d:]) // %s\n")
	case binary.I64Load16S:
		c.emitLoad(instr, opname, "stack[%d] = uint64(int64(int16(binary.LittleEndian.Uint16(m.memory[stack[%d] + %d:])))) // %s\n")
	case binary.I64Load16U:
		c.emitLoad(instr, opname, "stack[%d] = uint64(binary.LittleEndian.Uint16(m.memory[stack[%d] + %d:])) // %s\n")
	case binary.I64Load32S:
		c.emitLoad(instr, opname, "stack[%d] = uint64(int64(int32(binary.LittleEndian.Uint32(m.memory[stack[%d] + %d:])))) // %s\n")
	case binary.I64Load32U:
		c.emitLoad(instr, opname, "stack[%d] = uint64(binary.LittleEndian.Uint32(m.memory[stack[%d] + %d:])) // %s\n")
	case binary.I32Store, binary.F32Store:
		c.emitStore(instr, opname, "binary.LittleEndian.PutUint32(m.memory[stack[%d] + %d:], uint32(stack[%d])) // %s\n")
	case binary.I64Store, binary.F64Store:
		c.emitStore(instr, opname, "binary.LittleEndian.PutUint64(m.memory[stack[%d] + %d:], stack[%d]) // %s\n")
	case binary.I32Store8, binary.I64Store8:
		c.emitStore(instr, opname, "m.memory[stack[%d] + %d:] = byte(stack[%d]) // %s\n")
	case binary.I32Store16, binary.I64Store16:
		c.emitStore(instr, opname, "binary.LittleEndian.PutUint16(m.memory[stack[%d] + %d:], uint16(stack[%d])) // %s\n")
	case binary.I64Store32:
		c.emitStore(instr, opname, "binary.LittleEndian.PutUint32(m.memory[stack[%d] + %d:], uint32(stack[%d])) // %s\n")
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
		c.printf("stack[%d] = b2i(uint32(stack[%d]) == 0) // %s\n",
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
		c.printf("stack[%d] = b2i(stack[%d] == 0) // %s\n",
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
		c.printf("stack[%d] = uint64(bits.LeadingZeros32(stack[%d])) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32Ctz:
		c.printf("stack[%d] = uint64(bits.TrailingZeros32(stack[%d])) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32PopCnt:
		c.printf("stack[%d] = uint64(bits.OnesCount32(stack[%d])) // %s\n",
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
		c.printf("stack[%d] = uint32(stack[%d]) << (uint32(stack[%d]) %% 32) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I32ShrS:
		c.printf("stack[%d] = int32(uint32(stack[%d])) >> (uint32(stack[%d]) %% 32) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I32ShrU:
		c.printf("stack[%d] = uint32(stack[%d]) >> (uint32(stack[%d]) %% 32) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I32Rotl:
		c.printf("stack[%d] = bits.RotateLeft32(uint32(stack[%d]), int(uint32(stack[%d]))) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I32Rotr:
		c.printf("stack[%d] = bits.RotateLeft32(uint32(stack[%d]), -int(uint32(stack[%d]))) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I64Clz:
		c.printf("stack[%d] = uint64(bits.LeadingZeros64(stack[%d])) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64Ctz:
		c.printf("stack[%d] = uint64(bits.TrailingZeros64(stack[%d])) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64PopCnt:
		c.printf("stack[%d] = uint64(bits.OnesCount64(stack[%d])) // %s\n",
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
		c.printf("stack[%d] = stack[%d] << (stack[%d] %% 64) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I64ShrS:
		c.printf("stack[%d] = int64(stack[%d]) >> (stack[%d] %% 64) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I64ShrU:
		c.printf("stack[%d] = stack[%d] >> (stack[%d] %% 64) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I64Rotl:
		c.printf("stack[%d] = bits.RotateLeft64(stack[%d], int(stack[%d])) // %s\n",
			c.stackPtr-2, c.stackPtr-2, c.stackPtr-1, opname)
		c.stackPop()
	case binary.I64Rotr:
		c.printf("stack[%d] = bits.RotateLeft64(stack[%d], int(stack[%d])) // %s\n",
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
		c.printf("stack[%d] = uint64(uint32(stack[%d])) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32TruncF32S:
		c.printf("stack[%d] = uint64(uint32(int32(math.Trunc(float64(f32(stack[%d])))))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32TruncF32U:
		c.printf("stack[%d] = uint64(uint32(math.Trunc(float64(f32(stack[%d]))))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32TruncF64S:
		c.printf("stack[%d] = uint64(uint32(int32(math.Trunc(f64(stack[%d]))))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32TruncF64U:
		c.printf("stack[%d] = uint64(uint32(math.Trunc(f64(stack[%d])))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64ExtendI32S:
		c.printf("stack[%d] = uint64(int64(int32(uint32(stack[%d])))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64ExtendI32U:
		c.printf("stack[%d] = uint64(uint32(stack[%d])) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64TruncF32S:
		c.printf("stack[%d] = uint64(int64(math.Trunc(float64(f32(stack[%d]))))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64TruncF32U:
		c.printf("stack[%d] = uint64(math.Trunc(float64(f32(stack[%d])))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64TruncF64S:
		c.printf("stack[%d] = uint64(int64(math.Trunc(f64(stack[%d])))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I64TruncF64U:
		c.printf("stack[%d] = uint64(math.Trunc(f64(stack[%d]))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F32ConvertI32S:
		c.printf("stack[%d] = u32(float32(int32(uint32(stack[%d])))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F32ConvertI32U:
		c.printf("stack[%d] = u32(float32(uint32(stack[%d]))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F32ConvertI64S:
		c.printf("stack[%d] = u32(float32(int64(stack[%d]))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F32ConvertI64U:
		c.printf("stack[%d] = u32(float32(stack[%d])) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F32DemoteF64:
		c.printf("stack[%d] = u32(float32(f64(stack[%d]))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F64ConvertI32S:
		c.printf("stack[%d] = u64(float64(int32(uint32(stack[%d])))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F64ConvertI32U:
		c.printf("stack[%d] = u64(float64(uint32(stack[%d]))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F64ConvertI64S:
		c.printf("stack[%d] = u64(float64(int64(stack[%d]))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F64ConvertI64U:
		c.printf("stack[%d] = u64(float64(stack[%d])) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.F64PromoteF32:
		c.printf("stack[%d] = u64(float64(f32(stack[%d]))) // %s\n",
			c.stackPtr-1, c.stackPtr-1, opname)
	case binary.I32ReinterpretF32:
		c.printf("// %s\n", opname) // TODO
	case binary.I64ReinterpretF64:
		c.printf("// %s\n", opname) // TODO
	case binary.F32ReinterpretI32:
		c.printf("// %s\n", opname) // TODO
	case binary.F64ReinterpretI64:
		c.printf("// %s\n", opname) // TODO
	default:
		c.printf("// %s ???", opname)
	}
}

/*
l0: for {
	... // break
	break
}
*/
func (c *funcCompiler) emitBlock(expr []binary.Instruction, isLoop, hasResult bool) {
	c.enterBlock(isLoop, hasResult)
	c.printIndents()
	c.printf("_l%d: for {\n", c.blockDepth()-1)
	for _, instr := range expr {
		c.emitInstr(instr)
	}
	c.printIndents()
	c.printf("break } // end of _l%d\n", c.blockDepth()-1)
	c.exitBlock()
}

/*
l0: for {
	... // continue
	break
}
*/
func (c *funcCompiler) emitLoop() {
	panic("TODO")
}

/*
l0: for {
	if <cond> {
		...
		break
	} else {
		...
		break
	}
}
*/
func (c *funcCompiler) emitIf() {
	panic("TODO")
}
func (c *funcCompiler) emitBr(labelIdx uint32) {
	n := len(c.blocks) - int(labelIdx) - 1
	if c.blocks[n].isLoop {
		c.printf("continue _l%d // br\n", n)
	} else {
		c.printf("break _l%d // br\n", n)
	}
}
func (c *funcCompiler) emitBrIf(labelIdx uint32) {
	n := len(c.blocks) - int(labelIdx) - 1
	ret := "" // TODO: return
	br := "break"
	if c.blocks[n].isLoop {
		br = "continue"
	}
	c.printf("if stack[%d] != 0 { %s%s _l%d } // br_if\n",
		c.stackPtr-1, ret, br, n)
	c.stackPop()
}
func (c *funcCompiler) emitBrTable() {
	panic("TODO")
}
func (c *funcCompiler) emitReturn() {
	panic("TODO")
}
func (c *funcCompiler) emitCall(funcIdx uint32, opname string) {
	name, ft := getFuncNameAndType(c.module, int(funcIdx))
	paramCount := len(ft.ParamTypes)

	c.stackPtr -= paramCount
	if len(ft.ResultTypes) > 0 {
		c.printf("stack[%d] = ", c.stackPtr)
	}
	c.printf("m.f%d(", funcIdx)
	for i := 0; i < paramCount; i++ {
		if i > 0 {
			c.print(", ")
		}
		c.printf("stack[%d]", c.stackPtr+i)
	}
	if len(ft.ResultTypes) > 0 {
		c.stackPtr++
	}
	c.printf(") // %s %s\n", opname, name)
}
func (c *funcCompiler) emitCallIndirect() {
	panic("TODO")
}

func (c *funcCompiler) emitLoad(instr binary.Instruction, opname, tmpl string) {
	// tmpl = stack[%d] = binary.LittleEndian.Uint32(m.memory[stack[%d] + %d:]) // %s\n"
	c.printf(tmpl, c.stackPtr-1, c.stackPtr-1, instr.Args.(binary.MemArg).Offset, opname)
}
func (c *funcCompiler) emitStore(instr binary.Instruction, opname, tmpl string) {
	// tmpl = "binary.LittleEndian.PutUint32(m.memory[stack[%d] + %d:], uint32(stack[%d])) // %s\n"
	c.printf(tmpl, c.stackPtr-2, instr.Args.(binary.MemArg).Offset, c.stackPtr-1, opname)
	c.stackPtr -= 2
}
func (c *funcCompiler) emitMemSize(opname string) {
	c.printf("stack[%d] = uint64(len(m.memory) / (64*1024)) // %s\n",
		c.stackPush(), opname)
}
func (c *funcCompiler) emitMemGrow(opname string) {
	f := "n := stack[%d]; "
	f += "stack[%d] = uint64(len(m.memory) / (64*1024)); "
	f += "m.memory = append(m.memory, make([]byte, n*64*1024)...) // %s\n"
	c.printf(f, c.stackPtr-1, c.stackPtr-1, opname)
}

func (c *funcCompiler) emitConst(val uint64, opname string, arg interface{}) {
	c.printf("stack[%d] = 0x%x // %s %v\n",
		c.stackPush(), val, opname, arg)
}

func (c *funcCompiler) emitI32BinCmpU(operator, opname string) {
	c.printf("stack[%d] = b2i(uint32(stack[%d]) %s uint32(stack[%d])) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *funcCompiler) emitI32BinCmpS(operator, opname string) {
	c.printf("stack[%d] = b2i(int32(uint32(stack[%d])) %s int32(uint32(stack[%d]))) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *funcCompiler) emitI32BinArithU(operator, opname string) {
	c.printf("stack[%d] = uint32(stack[%d]) %s uint32(stack[%d]) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *funcCompiler) emitI32BinArithS(operator, opname string) {
	c.printf("stack[%d] = int32(uint32(stack[%d])) %s int32(uint32(stack[%d])) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}

func (c *funcCompiler) emitI64BinCmpU(operator, opname string) {
	c.printf("stack[%d] = b2i(stack[%d] %s stack[%d]) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *funcCompiler) emitI64BinCmpS(operator, opname string) {
	c.printf("stack[%d] = b2i(int64(stack[%d]) %s int64(stack[%d])) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *funcCompiler) emitI64BinArithU(operator, opname string) {
	c.printf("stack[%d] = stack[%d] %s stack[%d] // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *funcCompiler) emitI64BinArithS(operator, opname string) {
	c.printf("stack[%d] = int64(stack[%d]) %s int64(stack[%d]) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}

func (c *funcCompiler) emitF32BinCmp(operator, opname string) {
	c.printf("stack[%d] = b2i(f32(stack[%d])) %s f32(stack[%d]))) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *funcCompiler) emitF32BinArith(operator, opname string) {
	c.printf("stack[%d] = u32(f32(stack[%d]) %s f32(stack[%d])) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *funcCompiler) emitF32UnFC(funcName, opname string) {
	c.printf("stack[%d] = u32(float32(%s(float64(f32(stack[%d]))))) // %s\n",
		c.stackPtr-1, funcName, c.stackPtr-1, opname)
}
func (c *funcCompiler) emitF32BinFC(funcName, opname string) {
	c.printf("stack[%d] = u32(float32(%s(float64(f32(stack[%d])), float64(f32(stack[%d]))))) // %s\n",
		c.stackPtr-2, funcName, c.stackPtr-2, c.stackPtr-1, opname)
	c.stackPop()
}

func (c *funcCompiler) emitF64BinCmp(operator, opname string) {
	c.printf("stack[%d] = b2i(f64(stack[%d])) %s f64(stack[%d]))) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *funcCompiler) emitF64BinArith(operator, opname string) {
	c.printf("stack[%d] = u64(f64(stack[%d]) %s f64(stack[%d])) // %s\n",
		c.stackPtr-2, c.stackPtr-2, operator, c.stackPtr-1, opname)
	c.stackPop()
}
func (c *funcCompiler) emitF64UnFC(funcName, opname string) {
	c.printf("stack[%d] = u64(%s(f64(stack[%d]))) // %s\n",
		c.stackPtr-1, funcName, c.stackPtr-1, opname)
}
func (c *funcCompiler) emitF64BinFC(funcName, opname string) {
	c.printf("stack[%d] = u64(%s(f64(stack[%d]), f64(stack[%d]))) // %s\n",
		c.stackPtr-2, funcName, c.stackPtr-2, c.stackPtr-1, opname)
	c.stackPop()
}
