package binary

import "fmt"

const (
	MutConst byte = 0
	MutVar   byte = 1
)

type GlobalType struct {
	ValType ValType
	Mut     byte
}

func readGlobalType(reader *WasmReader) (gt GlobalType, err error) {
	if gt.ValType, err = readValType(reader); err != nil {
		return
	}
	if gt.Mut, err = reader.readByte(); err != nil {
		return
	}

	switch gt.Mut {
	case MutConst:
	case MutVar:
	default:
		err = fmt.Errorf("invalid mut: %d", gt.Mut)
	}
	return
}

func (gt GlobalType) String() string {
	return fmt.Sprintf("{type: %s, mut: %v}",
		ValTypeToStr(gt.ValType), gt.Mut == 1)
}
