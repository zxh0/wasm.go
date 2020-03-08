package interpreter

import "github.com/zxh0/wasm.go/binary"

const (
	btBlock = 0
	btLoop  = 1
	btFunc  = 2
)

type blockFrame struct {
	instrs []binary.Instruction
	rt     binary.BlockType // result type
	bt     byte             // block type
	bp     int              // operand stack base pointer
	pc     int              // program counter
}

type blockStack struct {
	frames []*blockFrame
}

func newBlockFrame(instrs []binary.Instruction, rt binary.BlockType,
	bt byte, bp int) *blockFrame {

	return &blockFrame{
		instrs: instrs,
		rt:     rt,
		bt:     bt,
		bp:     bp,
		pc:     0,
	}
}

func (bs *blockStack) blockDepth() int {
	return len(bs.frames)
}

func (bs *blockStack) topBlockFrame() *blockFrame {
	return bs.frames[len(bs.frames)-1]
}
func (bs *blockStack) topFuncFrame() *blockFrame {
	for n := len(bs.frames) - 1; n >= 0; n-- {
		if bf := bs.frames[n]; bf.bt == btFunc {
			return bf
		}
	}
	return nil
}

func (bs *blockStack) pushBlockFrame(bf *blockFrame) {
	bs.frames = append(bs.frames, bf)
}
func (bs *blockStack) popBlockFrame() *blockFrame {
	n := len(bs.frames)
	bf := bs.frames[n-1]
	bs.frames = bs.frames[:n-1]
	return bf
}
