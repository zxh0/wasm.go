package binary

type (
	TypeIdx   = uint32
	FuncIdx   = uint32
	TableIdx  = uint32
	MemIdx    = uint32
	GlobalIdx = uint32
	LocalIdx  = uint32
	LabelIdx  = uint32
)

func readIndices(reader *WasmReader) (vec []uint32, err error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	vec = make([]uint32, n)
	for i := range vec {
		if vec[i], err = reader.readVarU32(); err != nil {
			return
		}
	}

	return
}
