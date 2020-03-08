package main

import (
	"fmt"

	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
)

const DEBUG = false

func newSpecTestInstance() instance.Instance {
	_print := func(args ...interface{}) (interface{}, error) {
		if DEBUG {
			for _, arg := range args {
				fmt.Printf("spectest> %v\n", arg)
			}
		}
		return nil, nil
	}

	specTest := newNativeInstance()
	specTest.RegisterNoResultsFunc("print", _print)
	specTest.RegisterNoResultsFunc("print_i32", _print, binary.ValTypeI32)
	specTest.RegisterNoResultsFunc("print_i64", _print, binary.ValTypeI64)
	specTest.RegisterNoResultsFunc("print_f32", _print, binary.ValTypeF32)
	specTest.RegisterNoResultsFunc("print_f64", _print, binary.ValTypeF64)
	specTest.RegisterNoResultsFunc("print_i32_f32", _print, binary.ValTypeI32, binary.ValTypeF32)
	specTest.RegisterNoResultsFunc("print_f64_f64", _print, binary.ValTypeF64, binary.ValTypeF64)
	specTest.RegisterGlobal("global_i32", binary.ValTypeI32, false, 666)
	specTest.RegisterGlobal("global_f32", binary.ValTypeF32, false, 0)
	specTest.RegisterGlobal("global_f64", binary.ValTypeF64, false, 0)
	specTest.RegisterTable("table", 10, 20) // TODO
	specTest.RegisterMem("memory", 1, 2)    // TODO
	return specTest
}
