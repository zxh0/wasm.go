package main

import (
	"fmt"
	"reflect"

	"github.com/zxh0/wasm.go/instance"
)

var _ instance.Instance = (*NativeInstance)(nil)

type NativeInstance struct {
	mem instance.Memory
}

func (n *NativeInstance) GetGlobalValue(name string) (interface{}, error) {
	panic("implement me")
}

func (n *NativeInstance) CallFunc(name string, args ...interface{}) (interface{}, error) {
	panic("implement me")
}

func (n *NativeInstance) Get(name string) interface{} {
	//switch name + ft.GetSignature() {
	//case "print_str(i32,i32)->()":
	//	return n.printStr
	//case "print_i32(i32)->()",
	//	"print_i64(i64)->()",
	//	"print_f32(f32)->()",
	//	"print_f64(f64)->()":
	//	return printNum
	//case "assert_true(i32)->()":
	//	return assertTrue
	//case "assert_false(i32)->()":
	//	return assertFalse
	//case "assert_eq_i32(i32,i32)->()",
	//	"assert_eq_i64(i64,i64)->()",
	//	"assert_eq_f32(f32,f32)->()",
	//	"assert_eq_f64(f64,f64)->()":
	//	return assertEq
	//default:
	//	return nil
	//}
	panic("TODO")
}

func (n *NativeInstance) printStr(args ...interface{}) interface{} {
	offset := args[0].(int32)
	length := args[1].(int32)
	buf := make([]byte, length)
	n.mem.Read(uint64(offset), buf)
	fmt.Print(string(buf))
	return nil
}

func printNum(args ...interface{}) interface{} {
	fmt.Printf("%v\n", args[0])
	return nil
}

func assertTrue(args ...interface{}) interface{} {
	if a := args[0].(int32); a == 0 {
		panic(fmt.Errorf("not true: %d", a))
	}
	return nil
}
func assertFalse(args ...interface{}) interface{} {
	if a := args[0].(int32); a != 0 {
		panic(fmt.Errorf("not false: %d", a))
	}
	return nil
}
func assertEq(args ...interface{}) interface{} {
	a, b := args[0], args[1]
	if !reflect.DeepEqual(a, b) {
		panic(fmt.Errorf("%v != %v", a, b))
	}
	return nil
}
