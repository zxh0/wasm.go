// +build !jit

package jit

import (
	"fmt"

	"github.com/zxh0/wasm.go/binary"
)

func Compile(module binary.Module) {
	fmt.Println("build tag is not set: jit")
}
