package aot

import (
	"fmt"
	"github.com/zxh0/wasm.go/binary"
)

func Compile(module binary.Module) {
	c := &moduleCompiler{
		printer:    newPrinter(),
		moduleInfo: newModuleInfo(module),
	}
	c.compile()
	fmt.Println(c.sb.String())
}
