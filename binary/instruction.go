package binary

import "fmt"

type Expr = []Instruction

type Instruction struct {
	Opcode byte
	Args   interface{}
}

// block & loop
type BlockArgs struct {
	RT     BlockType
	Instrs []Instruction
}

type IfArgs struct {
	RT      BlockType
	Instrs1 []Instruction
	Instrs2 []Instruction
}

type BrTableArgs struct {
	Labels  []LabelIdx
	Default LabelIdx
}

type MemArg struct {
	Align  uint32
	Offset uint32
}

func readExpr(reader *WasmReader) (Expr, error) {
	instrs, end, err := readInstructions(reader)
	if err != nil {
		return nil, err
	}
	if end != _End {
		return nil, fmt.Errorf("invalid expr end: %d", end)
	}
	return instrs, nil
}

func readInstructions(reader *WasmReader) (instrs []Instruction, end byte, err error) {
	for {
		if b := reader.nextByte(); b == _Else || b == _End {
			end, err = reader.readByte()
			return
		}

		var instr Instruction
		if instr, err = readInstruction(reader); err != nil {
			return
		}
		instrs = append(instrs, instr)
	}
}

func readInstruction(reader *WasmReader) (instr Instruction, err error) {
	if instr.Opcode, err = reader.readByte(); err != nil {
		return
	}
	if opnames[instr.Opcode] == "" {
		err = fmt.Errorf("undefined opcode: 0x%02x", instr.Opcode)
		return
	}
	instr.Args, err = readArgs(reader, instr.Opcode)
	return
}

func readArgs(reader *WasmReader, opcode byte) (interface{}, error) {
	switch opcode {
	case Block, Loop:
		return readBlockArgs(reader)
	case If:
		return readIfArgs(reader)
	case Br, BrIf:
		return reader.readVarU32() // label_idx
	case BrTable:
		return readBrTableArgs(reader)
	case Call:
		return reader.readVarU32() // func_idx
	case CallIndirect:
		return readCallIndirectArgs(reader)
	case LocalGet, LocalSet, LocalTee:
		return reader.readVarU32() // local_idx
	case GlobalGet, GlobalSet:
		return reader.readVarU32() // global_idx
	case MemorySize, MemoryGrow:
		return nil, readZero(reader)
	case I32Const:
		return reader.readVarS32()
	case I64Const:
		return reader.readVarS64()
	case F32Const:
		return reader.readF32()
	case F64Const:
		return reader.readF64()
	default:
		if opcode >= I32Load && opcode <= I64Store32 {
			return readMemArg(reader)
		}
		return nil, nil
	}
}

func readBlockArgs(reader *WasmReader) (args BlockArgs, err error) {
	if args.RT, err = readBlockType(reader); err != nil {
		return
	}
	var end byte
	if args.Instrs, end, err = readInstructions(reader); err != nil {
		return
	}
	if end != _End {
		err = fmt.Errorf("invalid block end: %d", end)
		return
	}
	return
}

func readIfArgs(reader *WasmReader) (args IfArgs, err error) {
	if args.RT, err = readBlockType(reader); err != nil {
		return
	}
	var end byte
	if args.Instrs1, end, err = readInstructions(reader); err != nil {
		return
	}
	if end == _Else {
		if args.Instrs2, end, err = readInstructions(reader); err != nil {
			return
		}
		if end != _End {
			err = fmt.Errorf("invalid block end: %d", end)
			return
		}
	}
	return
}

func readBrTableArgs(reader *WasmReader) (args BrTableArgs, err error) {
	if args.Labels, err = readIndices(reader); err != nil {
		return
	}
	args.Default, err = reader.readVarU32()
	return
}

func readCallIndirectArgs(reader *WasmReader) (typeIdx uint32, err error) {
	if typeIdx, err = reader.readVarU32(); err != nil {
		return
	}
	err = readZero(reader)
	return
}

func readMemArg(reader *WasmReader) (memArg MemArg, err error) {
	if memArg.Align, err = reader.readVarU32(); err != nil {
		return
	}
	memArg.Offset, err = reader.readVarU32()
	return
}

func readZero(reader *WasmReader) error {
	b, err := reader.readByte()
	if err != nil {
		return err
	}
	if b != 0 {
		return fmt.Errorf("expected 0, got %d", b)
	}
	return nil
}

func (instr Instruction) GetOpname() string {
	return opnames[instr.Opcode]
}
func (instr Instruction) String() string {
	return opnames[instr.Opcode]
}
