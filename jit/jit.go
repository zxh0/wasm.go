// +build jit

package jit

import (
	"fmt"

	"github.com/tinygo-org/go-llvm"
)

func Test() {
	fmt.Printf("%v", llvm.LittleEndian)
}