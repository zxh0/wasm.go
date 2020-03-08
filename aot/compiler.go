package aot

import (
	"fmt"
	"strings"

	"github.com/zxh0/wasm.go/binary"
)

func Compile(module binary.Module) {
	c := &moduleCompiler{
		printer: printer{sb: &strings.Builder{}},
		module:  module,
	}
	c.compile()
	fmt.Println(c.sb.String())
}
