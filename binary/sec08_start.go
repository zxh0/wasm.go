package binary

//type StartSec = FuncIdx

func readStartSec(reader *WasmReader) (*uint32, error) {
	idx, err := reader.readVarU32()
	return &idx, err
}
