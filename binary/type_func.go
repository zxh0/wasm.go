package binary

import (
	"fmt"
	"strings"
)

type FuncType struct {
	ParamTypes  []ValType
	ResultTypes []ValType
}

func readFuncType(reader *WasmReader) (ft FuncType, err error) {
	var tag byte
	if tag, err = reader.readByte(); err != nil {
		return
	}
	if tag != 0x60 {
		err = fmt.Errorf("invalid functype tag: %d", tag)
		return
	}
	if ft.ParamTypes, err = readValTypes(reader); err != nil {
		return
	}
	ft.ResultTypes, err = readValTypes(reader)
	return
}

func (ft FuncType) Equal(ft2 FuncType) bool {
	//return reflect.DeepEqual(ft, ft2)
	if len(ft.ParamTypes) != len(ft2.ParamTypes) {
		return false
	}
	if len(ft.ResultTypes) != len(ft2.ResultTypes) {
		return false
	}
	for i, vt := range ft.ParamTypes {
		if vt != ft2.ParamTypes[i] {
			return false
		}
	}
	for i, vt := range ft.ResultTypes {
		if vt != ft2.ResultTypes[i] {
			return false
		}
	}
	return true
}

func (ft FuncType) String() string {
	return ft.GetSignature()
}

// (i32,i32)->(i32)
func (ft FuncType) GetSignature() string {
	sb := strings.Builder{}
	sb.WriteString("(")
	for i, vt := range ft.ParamTypes {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(ValTypeToStr(vt))
	}
	sb.WriteString(")->(")
	for i, vt := range ft.ResultTypes {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(ValTypeToStr(vt))
	}
	sb.WriteString(")")
	return sb.String()
}
