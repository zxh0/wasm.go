package spectest

import (
	"fmt"
	"math"
	"strings"

	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/instance"
	"github.com/zxh0/wasm.go/text"
)

type wastTester struct {
	script    *text.Script
	wasmImpl  WasmImpl
	instances map[string]instance.Instance
	instance  instance.Instance
}

func TestWast(script *text.Script) error {
	return newWastTester(script).test()
}

func newWastTester(script *text.Script) *wastTester {
	return &wastTester{
		script:   script,
		wasmImpl: WasmInterpreter{},
		instances: map[string]instance.Instance{
			"spectest": newSpecTestInstance(),
		},
	}
}

func (t *wastTester) test() (err error) {
	for _, cmd := range t.script.Cmds {
		switch x := cmd.(type) {
		case *text.WatModule:
			err = t.instantiate(x)
		case *text.BinaryModule:
			t.instantiateBin(x.Data)
		case *text.QuotedModule:
			panic("TODO")
		case *text.Register:
			t.instances[x.ModuleName] = t.instance
		case *text.Action:
			_, _ = t.runAction(x) // TODO
		case *text.Assertion:
			err = t.runAssertion(x)
		case text.Meta:
			panic("TODO")
		default:
			panic("unreachable")
		}
		if err != nil {
			return
		}
	}
	return
}

func (t *wastTester) instantiate(m *text.WatModule) (err error) {
	t.instance, err = t.wasmImpl.Instantiate(*m.Module, t.instances)
	if err == nil && m.Name != "" {
		t.instances[m.Name] = t.instance
	}
	return err
}
func (t *wastTester) instantiateBin(data []byte) {
	_, err := t.wasmImpl.InstantiateBin(data, t.instances)
	if err != nil {
		panic(err)
	}
	// TODO: check
}

func (t *wastTester) runAssertion(a *text.Assertion) error {
	switch a.Kind {
	case text.AssertReturn:
		result, err := t.runAction(a.Action)
		return assertReturn(a.Result, result, err)
	case text.AssertTrap:
		if a.Action != nil {
			result, err := t.runAction(a.Action)
			return assertTrap(a.Failure, result, err)
		} else {
			err := t.instantiate(a.Module.(*text.WatModule))
			return assertTrap(a.Failure, err, err)
		}
	case text.AssertExhaustion:
		// panic("TODO")
	case text.AssertMalformed:
		switch m := a.Module.(type) {
		case *text.BinaryModule:
			_, err := binary.Decode(m.Data)
			if a.Failure != "length out of bounds" { // TODO
				return assertError(a.Failure, err)
			}
		case *text.QuotedModule:
			// panic("TODO")
		}
	case text.AssertInvalid:
		m := *(a.Module.(*text.WatModule).Module)
		err := t.wasmImpl.Validate(m)
		return assertError(a.Failure, err)
	case text.AssertUnlinkable:
		err := t.instantiate(a.Module.(*text.WatModule))
		return assertError(a.Failure, err)
	default:
		panic("TODO")
	}
	return nil
}

func (t *wastTester) runAction(a *text.Action) (interface{}, error) {
	_i := t.instance
	if a.ModuleName != "" {
		_i = t.instances[a.ModuleName]
	}

	switch a.Kind {
	case text.ActionInvoke:
		//println("invoke " + a.ItemName)
		return _i.CallFunc(a.ItemName, getConsts(a.Expr)...)
	case text.ActionGet:
		//println("get " + a.ItemName)
		return _i.GetGlobalValue(a.ItemName)
	default:
		panic("unreachable")
	}
}

func assertReturn(expected []binary.Instruction,
	result interface{}, err error) error {

	if err != nil {
		result = err
	}

	expectedVals := getConsts(expected)
	var expectedVal interface{} = nil
	if n := len(expectedVals); n == 1 {
		expectedVal = expectedVals[0]
	} else if n > 1 {
		panic("TODO")
	}

	if isNaN32(expectedVal) { // TODO
		if !isNaN32(result) {
			return fmt.Errorf("expected return: NaN, got: %v", result)
		}
	} else if isNaN64(expectedVal) { // TODO
		if !isNaN64(result) {
			return fmt.Errorf("expected return: NaN, got: %v", result)
		}
	} else if result != expectedVal {
		return fmt.Errorf("expected return: %v, got: %v", expectedVal, result)
	}
	return nil
}
func assertTrap(expectedErr string, result interface{}, err error) error {
	if err == nil {
		return fmt.Errorf("expected trap: %v, got: %v",
			expectedErr, result)
	}
	if strings.Index(err.Error(), expectedErr) < 0 {
		return fmt.Errorf("expected trap: %v, got: %v",
			expectedErr, err)
	}
	return nil
}
func assertError(expectedErr string, err error) error {
	if err == nil || strings.Index(err.Error(), expectedErr) < 0 {
		return fmt.Errorf("expected: %v, got: %v",
			expectedErr, err)
	}
	return nil
}

func getConsts(expr []binary.Instruction) []interface{} {
	args := make([]interface{}, len(expr))
	for i, instr := range expr {
		switch instr.Opcode {
		case binary.I32Const:
			args[i] = instr.Args.(int32)
		case binary.I64Const:
			args[i] = instr.Args.(int64)
		case binary.F32Const:
			args[i] = instr.Args.(float32)
		case binary.F64Const:
			args[i] = instr.Args.(float64)
		default:
			panic("TODO")
		}
	}
	return args
}

func isNaN32(x interface{}) bool {
	f, ok := x.(float32)
	return ok && math.IsNaN(float64(f))
}
func isNaN64(x interface{}) bool {
	f, ok := x.(float64)
	return ok && math.IsNaN(f)
}
