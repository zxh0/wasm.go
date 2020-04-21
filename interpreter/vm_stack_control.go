package interpreter

import (
	"github.com/zxh0/wasm.go/binary"
)

const (
	maxCallDepth = 65535
)

type controlFrame struct {
	opcode byte
	bt     binary.FuncType      // block type
	instrs []binary.Instruction // expr
	bp     int                  // base pointer (operand stack)
	pc     int                  // program counter
}

type controlStack struct {
	frames    []*controlFrame
	callDepth int
}

func newControlFrame(opcode byte,
	bt binary.FuncType, instrs []binary.Instruction, bp int) *controlFrame {

	return &controlFrame{
		opcode: opcode,
		bt:     bt,
		instrs: instrs,
		bp:     bp,
		pc:     0,
	}
}

func (cs *controlStack) reset() {
	cs.frames = cs.frames[0:]
	cs.callDepth = 0
}
func (cs *controlStack) controlDepth() int {
	return len(cs.frames)
}

func (cs *controlStack) topControlFrame() *controlFrame {
	return cs.frames[len(cs.frames)-1]
}
func (cs *controlStack) topCallFrame() (*controlFrame, int) {
	for n := len(cs.frames) - 1; n >= 0; n-- {
		if cf := cs.frames[n]; cf.opcode == binary.Call {
			return cf, len(cs.frames) - 1 - n
		}
	}
	return nil, -1
}

func (cs *controlStack) pushControlFrame(cf *controlFrame) {
	cs.frames = append(cs.frames, cf)
	if cf.opcode == binary.Call {
		cs.callDepth++
		if cs.callDepth > maxCallDepth {
			panic(errCallStackOverflow)
		}
	}
}
func (cs *controlStack) popControlFrame() *controlFrame {
	cf := cs.frames[len(cs.frames)-1]
	cs.frames = cs.frames[:len(cs.frames)-1]
	if cf.opcode == binary.Call {
		cs.callDepth--
	}
	return cf
}
