package binary

import (
	"fmt"
)

type Limits struct {
	Tag byte
	Min uint32
	Max uint32
}

func readLimits(reader *WasmReader) (limits Limits, err error) {
	if limits.Tag, err = reader.readByte(); err != nil {
		return
	}
	if limits.Min, err = reader.readVarU32(); err != nil {
		return
	}
	if limits.Tag == 1 {
		limits.Max, err = reader.readVarU32()
	}
	return
}

func (limits Limits) String() string {
	return fmt.Sprintf("{min: %d, max: %d}",
		limits.Min, limits.Max)
}
