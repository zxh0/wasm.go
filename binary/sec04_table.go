package binary

//type TableSec = []TableType

func readTableSec(reader *WasmReader) (vec []TableType, err error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	vec = make([]TableType, n)
	for i := range vec {
		if vec[i], err = readTableType(reader); err != nil {
			return
		}
	}

	return
}
