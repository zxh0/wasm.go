package binary

type CustomSec struct {
	Name string
	// TODO
}

func readCustomSec(reader *WasmReader) (sec CustomSec, err error) {
	if sec.Name, err = reader.readName(); err != nil {
		return
	}

	// TODO
	reader.data = nil
	return
}
