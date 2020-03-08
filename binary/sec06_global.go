package binary

//type GlobalSec []Global

type Global struct {
	Type GlobalType
	Expr Expr
}

func readGlobalSec(reader *WasmReader) (vec []Global, err error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	vec = make([]Global, n)
	for i := range vec {
		global := Global{}
		if global.Type, err = readGlobalType(reader); err != nil {
			return
		}
		if global.Expr, err = readExpr(reader); err != nil {
			return
		}
		vec[i] = global
	}

	return
}
