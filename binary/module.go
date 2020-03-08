package binary

import (
	"fmt"
	"io/ioutil"
)

const (
	MagicNumber = 0x6D736100 // `\0asm`
	Version     = 0x00000001 // 1
)

const (
	SecCustomID = iota
	SecTypeID
	SecImportID
	SecFuncID
	SecTableID
	SecMemID
	SecGlobalID
	SecExportID
	SecStartID
	SecElemID
	SecCodeID
	SecDataID
)

type Module struct {
	Magic      uint32
	Version    uint32
	CustomSecs []CustomSec
	TypeSec    []FuncType
	ImportSec  []Import
	FuncSec    []TypeIdx
	TableSec   []TableType
	MemSec     []MemType
	GlobalSec  []Global
	ExportSec  []Export
	StartSec   *FuncIdx
	ElemSec    []Elem
	CodeSec    []Code
	DataSec    []Data
}

func DecodeFile(filename string) (Module, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return Module{}, err
	}
	return Decode(data)
}

func Decode(data []byte) (Module, error) {
	reader := WasmReader{data: data}
	return readModule(&reader)
}

// TODO: return *Module ?
func readModule(reader *WasmReader) (module Module, err error) {
	if module.Magic, err = reader.readU32(); err != nil {
		return
	}
	if module.Magic != MagicNumber {
		err = fmt.Errorf("invalid magic number: 0x%x", module.Magic)
		return
	}
	if module.Version, err = reader.readU32(); err != nil {
		return
	}
	if module.Version != Version {
		err = fmt.Errorf("unsupported version: %d", module.Version)
		return
	}
	err = readSections(reader, &module)
	return
}

func readSections(reader *WasmReader, module *Module) (err error) {
	lastSecID := byte(0)
	for reader.remaining() > 0 {
		var secID byte
		if secID, err = reader.readByte(); err != nil {
			return
		}
		if secID > SecCustomID {
			if secID < lastSecID {
				err = fmt.Errorf("invalid sec ID: %d", secID)
				return
			}
			lastSecID = secID
		}

		var secCont []byte
		if secCont, err = reader.readBytes(); err != nil {
			return
		}
		if err = decodeSec(secID, secCont, module); err != nil {
			return
		}
	}
	return
}

func decodeSec(secID byte, cont []byte, module *Module) (err error) {
	secReader := WasmReader{data: cont}
	if secID == SecCustomID {
		var sec CustomSec
		if sec, err = readCustomSec(&secReader); err != nil {
			return
		}
		module.CustomSecs = append(module.CustomSecs, sec)
	} else {
		if err = readNonCustomSec(secID, &secReader, module); err != nil {
			return
		}
	}
	if secReader.remaining() > 0 {
		err = fmt.Errorf("invalid sec, id=%d", secID)
	}
	return
}

func readNonCustomSec(secID byte, reader *WasmReader, module *Module) (err error) {
	switch secID {
	case SecTypeID:
		module.TypeSec, err = readTypeSec(reader)
	case SecImportID:
		module.ImportSec, err = readImportSec(reader)
	case SecFuncID:
		module.FuncSec, err = readIndices(reader)
	case SecTableID:
		module.TableSec, err = readTableSec(reader)
	case SecMemID:
		module.MemSec, err = readMemSec(reader)
	case SecGlobalID:
		module.GlobalSec, err = readGlobalSec(reader)
	case SecExportID:
		module.ExportSec, err = readExportSec(reader)
	case SecStartID:
		module.StartSec, err = readStartSec(reader)
	case SecElemID:
		module.ElemSec, err = readElemSec(reader)
	case SecCodeID:
		module.CodeSec, err = readCodeSec(reader)
	case SecDataID:
		module.DataSec, err = readDataSec(reader)
	}
	return
}
