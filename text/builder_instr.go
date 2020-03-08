package text

import "github.com/zxh0/wasm.go/binary"

func newInstruction(opname string) binary.Instruction {
	opcode, ok := binary.GetOpcode(opname)
	if !ok {
		panic("unreachable")
	}
	return binary.Instruction{Opcode: opcode}
}

func newI32Const0() binary.Instruction {
	return binary.Instruction{
		Opcode: binary.I32Const,
		Args:   int32(0),
	}
}

func newBlockInstr(opname string, rt binary.BlockType,
	expr1, expr2 []binary.Instruction) binary.Instruction {

	instr := newInstruction(opname)
	switch instr.Opcode {
	case binary.Block, binary.Loop:
		instr.Args = binary.BlockArgs{
			RT:     rt,
			Instrs: expr1,
		}
	case binary.If:
		ifArgs := binary.IfArgs{
			RT:      rt,
			Instrs1: expr1,
			Instrs2: expr2,
		}
		instr.Args = ifArgs
	default:
		panic("unreachable")
	}
	return instr
}

func checkArgs(instr binary.Instruction) {
	// TODO
}
