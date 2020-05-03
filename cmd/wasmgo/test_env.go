package main

import (
	"fmt"

	"github.com/zxh0/wasm.go/instance"
)

func newTestEnv() instance.Module {
	env := instance.NewNativeInstance()
	env.RegisterFunc("assert_true(i32)->()", assertTrue)
	env.RegisterFunc("assert_false(i32)->()", assertFalse)
	env.RegisterFunc("assert_eq_i32(i32,i32)->()", assertEqI32)
	env.RegisterFunc("assert_eq_i64(i64,i64)->()", assertEqI64)
	env.RegisterFunc("assert_eq_f32(f32,f32)->()", assertEqF32)
	env.RegisterFunc("assert_eq_f64(f64,f64)->()", assertEqF64)
	env.RegisterFunc("print_i32(i32)->()", printI32)
	env.RegisterFunc("print_i64(i64)->()", printI64)
	env.RegisterFunc("print_char(i32)->()", printChar)
	return env
}

func assertTrue(args []interface{}) ([]interface{}, error) {
	fmt.Printf("assert_true: %v\n", args)
	if args[0].(int32) == 1 {
		return nil, nil
	}
	panic(fmt.Errorf("not true: %v", args))
}
func assertFalse(args []interface{}) ([]interface{}, error) {
	fmt.Printf("assert_false: %v\n", args)
	if args[0].(int32) == 0 {
		return nil, nil
	}
	panic(fmt.Errorf("not false: %v", args))
}
func assertEqI32(args []interface{}) ([]interface{}, error) {
	fmt.Printf("assert_eq_i32: %v\n", args)
	if args[0].(int32) == args[1].(int32) {
		return nil, nil
	}
	panic(fmt.Errorf("not equal: %v", args))
}
func assertEqI64(args []interface{}) ([]interface{}, error) {
	fmt.Printf("assert_eq_i64: %v\n", args)
	if args[0].(int64) == args[1].(int64) {
		return nil, nil
	}
	panic(fmt.Errorf("not equal: %v", args))
}
func assertEqF32(args []interface{}) ([]interface{}, error) {
	fmt.Printf("assert_eq_f32: %v\n", args)
	if args[0].(float32) == args[1].(float32) {
		return nil, nil
	}
	panic(fmt.Errorf("not equal: %v", args))
}
func assertEqF64(args []interface{}) ([]interface{}, error) {
	fmt.Printf("assert_eq_f64: %v\n", args)
	if args[0].(float64) == args[1].(float64) {
		return nil, nil
	}
	panic(fmt.Errorf("not equal: %v", args))
}

func printI32(args []interface{}) ([]interface{}, error) {
	fmt.Printf("%v\n", args[0])
	return nil, nil
}
func printI64(args []interface{}) ([]interface{}, error) {
	fmt.Printf("%v\n", args[0])
	return nil, nil
}
func printChar(args []interface{}) ([]interface{}, error) {
	fmt.Printf("%c", args[0])
	return nil, nil
}
