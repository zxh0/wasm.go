package text

import (
	"github.com/zxh0/wasm.go/binary"
)

func newInstruction(opname string) binary.Instruction {
	if opcode, ok := binary.GetOpcode(opname); ok {
		return binary.Instruction{Opcode: opcode}
	}
	return newTruncSat(opname)
}
func newTruncSat(opname string) binary.Instruction {
	instr := binary.Instruction{Opcode: binary.TruncSat}
	switch opname {
	case "i32.trunc_sat_f32_s":
		instr.Args = byte(0x00)
	case "i32.trunc_sat_f32_u":
		instr.Args = byte(0x01)
	case "i32.trunc_sat_f64_s":
		instr.Args = byte(0x02)
	case "i32.trunc_sat_f64_u":
		instr.Args = byte(0x03)
	case "i64.trunc_sat_f32_s":
		instr.Args = byte(0x04)
	case "i64.trunc_sat_f32_u":
		instr.Args = byte(0x05)
	case "i64.trunc_sat_f64_s":
		instr.Args = byte(0x06)
	case "i64.trunc_sat_f64_u":
		instr.Args = byte(0x07)
	default:
		panic("unreachable")
	}
	return instr
}

func newI32Const0() binary.Instruction {
	return binary.Instruction{
		Opcode: binary.I32Const,
		Args:   int32(0),
	}
}

func newBlockInstr(opname string, bt binary.BlockType,
	expr1, expr2 []binary.Instruction) binary.Instruction {

	instr := newInstruction(opname)
	switch instr.Opcode {
	case binary.Block, binary.Loop:
		instr.Args = binary.BlockArgs{
			BT:     bt,
			Instrs: expr1,
		}
	case binary.If:
		ifArgs := binary.IfArgs{
			BT:      bt,
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
