package binary

//type MemSec = []MemType

func readMemSec(reader *WasmReader) (vec []MemType, err error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	vec = make([]MemType, n)
	for i := range vec {
		if vec[i], err = readLimits(reader); err != nil {
			return
		}
	}

	return
}
