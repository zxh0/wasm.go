package binary

const NoVal = 0x40

type BlockType = []ValType

func readBlockType(reader *WasmReader) ([]ValType, error) {
	if b, err := reader.readByte(); err != nil {
		return nil, err
	} else if b == NoVal {
		return nil, nil
	} else {
		if err := checkValType(b); err != nil {
			return nil, err
		}
		return []ValType{b}, nil
	}
}
