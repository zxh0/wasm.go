package main

import (
	"fmt"

	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
)

func newTestEnv() instance.Instance {
	env := instance.NewNativeInstance()
	env.RegisterFunc("assert_eq_i32", assertEqI32, binary.ValTypeI32, binary.ValTypeI32)
	return env
}

func assertEqI32(args ...interface{}) (interface{}, error) {
	fmt.Printf("assert_eq_i32: %v\n", args)
	if args[0].(int32) == args[1].(int32) {
		return nil, nil
	}
	panic(fmt.Errorf("not equal: %v", args))
}
