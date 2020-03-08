package binary

//type TypeSec = []FuncType

func readTypeSec(reader *WasmReader) (vec []FuncType, err error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	vec = make([]FuncType, n)
	for i := range vec {
		if vec[i], err = readFuncType(reader); err != nil {
			return
		}
	}

	return
}
