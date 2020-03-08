package binary

//type DataSec = []Data

type Data struct {
	Mem    MemIdx
	Offset Expr
	Init   []byte
}

func readDataSec(reader *WasmReader) (vec []Data, err error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	vec = make([]Data, n)
	for i := range vec {
		if vec[i], err = readData(reader); err != nil {
			return
		}
	}

	return
}

func readData(reader *WasmReader) (data Data, err error) {
	if data.Mem, err = reader.readVarU32(); err != nil {
		return
	}
	if data.Offset, err = readExpr(reader); err != nil {
		return
	}
	data.Init, err = reader.readBytes()
	return
}
