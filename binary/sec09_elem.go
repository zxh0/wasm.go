package binary

//type ElemSec = []Elem

type Elem struct {
	Table  TableIdx
	Offset Expr
	Init   []FuncIdx
}

func readElemSec(reader *WasmReader) (vec []Elem, err error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	vec = make([]Elem, n)
	for i := range vec {
		if vec[i], err = readElem(reader); err != nil {
			return
		}
	}

	return
}

func readElem(reader *WasmReader) (elem Elem, err error) {
	if elem.Table, err = reader.readVarU32(); err != nil {
		return
	}
	if elem.Offset, err = readExpr(reader); err != nil {
		return
	}
	elem.Init, err = readIndices(reader)
	return
}
