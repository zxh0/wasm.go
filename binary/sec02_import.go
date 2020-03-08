package binary

import "fmt"

const (
	ImportTagFunc   = 0
	ImportTagTable  = 1
	ImportTagMem    = 2
	ImportTagGlobal = 3
)

//type ImportSec = []Import

type Import struct {
	Module string
	Name   string
	Desc   ImportDesc
}

type ImportDesc struct {
	Tag      byte
	FuncType TypeIdx    // tag=0
	Table    TableType  // tag=1
	Mem      MemType    // tag=2
	Global   GlobalType // tag=3
}

func readImportSec(reader *WasmReader) (vec []Import, err error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	vec = make([]Import, n)
	for i := range vec {
		if vec[i], err = readImport(reader); err != nil {
			return
		}
	}
	return
}

func readImport(reader *WasmReader) (imp Import, err error) {
	if imp.Module, err = reader.readName(); err != nil {
		return
	}
	if imp.Name, err = reader.readName(); err != nil {
		return
	}
	imp.Desc, err = readImportDesc(reader)
	return
}

func readImportDesc(reader *WasmReader) (desc ImportDesc, err error) {
	if desc.Tag, err = reader.readByte(); err != nil {
		return
	}

	switch desc.Tag {
	case ImportTagFunc:
		desc.FuncType, err = reader.readVarU32()
	case ImportTagTable:
		desc.Table, err = readTableType(reader)
	case ImportTagMem:
		desc.Mem, err = readLimits(reader)
	case ImportTagGlobal:
		desc.Global, err = readGlobalType(reader)
	default:
		err = fmt.Errorf("invalid import desc tag: %d", desc.Tag)
	}
	return
}
