package aot

import (
	"fmt"

	"github.com/zxh0/wasm.go/binary"
)

func getImportedFuncCount(m binary.Module) int {
	n := 0
	for _, imp := range m.ImportSec {
		if imp.Desc.Tag == binary.ImportTagFunc {
			n++
		}
	}
	return n
}

func getFuncNameAndType(m binary.Module, funcIdx int) (string, binary.FuncType) {
	i := 0
	for _, imp := range m.ImportSec {
		if imp.Desc.Tag == binary.ImportTagFunc {
			if i == funcIdx {
				name := imp.Module + "." + imp.Name
				return name, m.TypeSec[imp.Desc.FuncType]
			}
			i++
		}
	}
	ftIdx := m.FuncSec[funcIdx-i]
	name := fmt.Sprintf("func#%d", funcIdx)
	return name, m.TypeSec[ftIdx]
}

func getMemPageMin(m binary.Module) int {
	if len(m.MemSec) > 0 {
		return int(m.MemSec[0].Min)
	}
	return 0
}
