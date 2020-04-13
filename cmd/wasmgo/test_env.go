package main

import (
	"fmt"

	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
)

func newTestEnv() instance.Instance {
	env := instance.NewNativeInstance()
	env.RegisterFunc("assert_true", assertTrue, binary.ValTypeI32, binary.NoVal)
	env.RegisterFunc("assert_false", assertFalse, binary.ValTypeI32, binary.NoVal)
	env.RegisterFunc("assert_eq_i32", assertEqI32, binary.ValTypeI32, binary.ValTypeI32, binary.NoVal)
	env.RegisterFunc("assert_eq_i64", assertEqI64, binary.ValTypeI64, binary.ValTypeI64, binary.NoVal)
	env.RegisterFunc("assert_eq_f32", assertEqF32, binary.ValTypeF32, binary.ValTypeF32, binary.NoVal)
	env.RegisterFunc("assert_eq_f64", assertEqF64, binary.ValTypeF64, binary.ValTypeF64, binary.NoVal)
	env.RegisterFunc("print_i32", printI32, binary.ValTypeI32, binary.NoVal)
	env.RegisterFunc("print_char", printChar, binary.ValTypeI32, binary.NoVal)
	return env
}

func assertTrue(args ...interface{}) ([]interface{}, error) {
	fmt.Printf("assert_true: %v\n", args)
	if args[0].(int32) == 1 {
		return nil, nil
	}
	panic(fmt.Errorf("not true: %v", args))
}
func assertFalse(args ...interface{}) ([]interface{}, error) {
	fmt.Printf("assert_false: %v\n", args)
	if args[0].(int32) == 0 {
		return nil, nil
	}
	panic(fmt.Errorf("not false: %v", args))
}
func assertEqI32(args ...interface{}) ([]interface{}, error) {
	fmt.Printf("assert_eq_i32: %v\n", args)
	if args[0].(int32) == args[1].(int32) {
		return nil, nil
	}
	panic(fmt.Errorf("not equal: %v", args))
}
func assertEqI64(args ...interface{}) ([]interface{}, error) {
	fmt.Printf("assert_eq_i64: %v\n", args)
	if args[0].(int64) == args[1].(int64) {
		return nil, nil
	}
	panic(fmt.Errorf("not equal: %v", args))
}
func assertEqF32(args ...interface{}) ([]interface{}, error) {
	fmt.Printf("assert_eq_f32: %v\n", args)
	if args[0].(float32) == args[1].(float32) {
		return nil, nil
	}
	panic(fmt.Errorf("not equal: %v", args))
}
func assertEqF64(args ...interface{}) ([]interface{}, error) {
	fmt.Printf("assert_eq_f64: %v\n", args)
	if args[0].(float64) == args[1].(float64) {
		return nil, nil
	}
	panic(fmt.Errorf("not equal: %v", args))
}

func printI32(args ...interface{}) ([]interface{}, error) {
	fmt.Printf("%v\n", args[0])
	return nil, nil
}
func printChar(args ...interface{}) ([]interface{}, error) {
	fmt.Printf("%c", args[0])
	return nil, nil
}
