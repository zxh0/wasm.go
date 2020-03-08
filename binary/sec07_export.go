package binary

import "fmt"

const (
	ExportTagFunc   = 0
	ExportTagTable  = 1
	ExportTagMem    = 2
	ExportTagGlobal = 3
)

//type ExportSec = []Export

type Export struct {
	Name string
	Desc ExportDesc
}

type ExportDesc struct {
	Tag byte
	Idx uint32
}

func readExportSec(reader *WasmReader) (vec []Export, err error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	vec = make([]Export, n)
	for i := range vec {
		if vec[i], err = readExport(reader); err != nil {
			return
		}
	}

	return
}

func readExport(reader *WasmReader) (exp Export, err error) {
	if exp.Name, err = reader.readName(); err != nil {
		return
	}
	exp.Desc, err = readExportDesc(reader)
	return
}

func readExportDesc(reader *WasmReader) (desc ExportDesc, err error) {
	if desc.Tag, err = reader.readByte(); err != nil {
		return
	}
	if desc.Idx, err = reader.readVarU32(); err != nil {
		return
	}
	switch desc.Tag {
	case ExportTagFunc: // func_idx
	case ExportTagTable: // table_idx
	case ExportTagMem: // mem_idx
	case ExportTagGlobal: // global_idx
	default:
		err = fmt.Errorf("invalid export desc tag: %d", desc.Tag)
	}
	return
}
