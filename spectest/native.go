package spectest

import (
	"fmt"

	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
	"github.com/zxh0/wasm.go/interpreter"
)

const Debug = false

func newSpecTestInstance() instance.Module {
	specTest := instance.NewNativeInstance()
	specTest.RegisterFunc("print()->()", _print)
	specTest.RegisterFunc("print_i32(i32)->()", _print)
	specTest.RegisterFunc("print_i64(i64)->()", _print)
	specTest.RegisterFunc("print_f32(f32)->()", _print)
	specTest.RegisterFunc("print_f64(f64)->()", _print)
	specTest.RegisterFunc("print_i32_f32(i32,f32)->()", _print)
	specTest.RegisterFunc("print_f64_f64(f64,f64)->()", _print)
	specTest.Register("global_i32", interpreter.NewGlobal(binary.ValTypeI32, false, 666))
	specTest.Register("global_f32", interpreter.NewGlobal(binary.ValTypeF32, false, 0))
	specTest.Register("global_f64", interpreter.NewGlobal(binary.ValTypeF64, false, 0))
	specTest.Register("table", interpreter.NewTable(10, 20))
	specTest.Register("memory", interpreter.NewMemory(1, 2))
	return specTest
}

func _print(args []interface{}) ([]interface{}, error) {
	if Debug {
		for _, arg := range args {
			fmt.Printf("spectest> %v\n", arg)
		}
	}
	return nil, nil
}
