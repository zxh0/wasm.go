package binary

import (
	"fmt"
)

const (
	ValTypeI32 ValType = 0x7F // i32
	ValTypeI64 ValType = 0x7E // i64
	ValTypeF32 ValType = 0x7D // f32
	ValTypeF64 ValType = 0x7C // f64
)

type ValType = byte

func readValTypes(reader *WasmReader) (vec []ValType, err error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	vec = make([]ValType, n)
	for i := range vec {
		if vec[i], err = readValType(reader); err != nil {
			return
		}
	}
	return
}

func readValType(reader *WasmReader) (vt ValType, err error) {
	if vt, err = reader.readByte(); err != nil {
		return
	}

	err = checkValType(vt)
	return
}

func checkValType(vt byte) error {
	switch vt {
	case ValTypeI32:
	case ValTypeI64:
	case ValTypeF32:
	case ValTypeF64:
	default:
		return fmt.Errorf("invalid valtype: %d", vt)
	}
	return nil
}

func ValTypeToStr(vt ValType) string {
	switch vt {
	case ValTypeI32:
		return "i32"
	case ValTypeI64:
		return "i64"
	case ValTypeF32:
		return "f32"
	case ValTypeF64:
		return "f64"
	default:
		panic(fmt.Errorf("invalid valtype: %d", vt))
	}
}
