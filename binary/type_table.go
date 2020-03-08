package binary

import "fmt"

const FuncRef = 0x70

//type ElemType = byte

type TableType struct {
	ElemType byte
	Limits   Limits
}

func readTableType(reader *WasmReader) (tt TableType, err error) {
	if tt.ElemType, err = reader.readByte(); err != nil {
		return
	}
	if tt.ElemType != FuncRef {
		err = fmt.Errorf("invalid elemtype: %d", tt.ElemType)
	}
	tt.Limits, err = readLimits(reader)
	return
}
