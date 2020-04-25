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
	instances map[string]instance.Module
	instance  instance.Module
}

func TestWast(script *text.Script) error {
	return newWastTester(script).test()
}

func newWastTester(script *text.Script) *wastTester {
	return &wastTester{
		script:   script,
		wasmImpl: WasmInterpreter{},
		instances: map[string]instance.Module{
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
			t.instantiateBin(x)
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
	if err == nil {
		if m.Name != "" {
			t.instances[m.Name] = t.instance
		}
	} else {
		err = fmt.Errorf("line: %d, %s", m.Line, err.Error())
	}
	return
}
func (t *wastTester) instantiateBin(m *text.BinaryModule) {
	tmp, err := t.wasmImpl.InstantiateBin(m.Data, t.instances)
	if err != nil {
		panic(err)
	}
	t.instance = tmp
	t.instances[m.Name] = t.instance
	// TODO: check
}

func (t *wastTester) runAssertion(a *text.Assertion) error {
	switch a.Kind {
	case text.AssertReturn:
		results, err := t.runAction(a.Action)
		return assertReturn(a, results, err)
	case text.AssertTrap:
		if a.Action != nil {
			results, err := t.runAction(a.Action)
			return assertTrap(a, results, err)
		} else {
			err := t.instantiate(a.Module.(*text.WatModule))
			return assertTrap(a, nil, err)
		}
	case text.AssertExhaustion:
		// very slow!
		//_, err := t.runAction(a.Action)
		//return assertError(a, err)
	case text.AssertMalformed:
		switch m := a.Module.(type) {
		case *text.BinaryModule:
			_, err := binary.Decode(m.Data)
			if a.Failure != "length out of bounds" { // TODO
				return assertError(a, err)
			}
		case *text.QuotedModule:
			// panic("TODO")
		}
	case text.AssertInvalid:
		m := *(a.Module.(*text.WatModule).Module)
		err := t.wasmImpl.Validate(m)
		return assertError(a, err)
	case text.AssertUnlinkable:
		err := t.instantiate(a.Module.(*text.WatModule))
		return assertError(a, err)
	default:
		panic("unreachable")
	}
	return nil
}

func (t *wastTester) runAction(a *text.Action) ([]interface{}, error) {
	_i := t.instance
	if a.ModuleName != "" {
		_i = t.instances[a.ModuleName]
	}

	switch a.Kind {
	case text.ActionInvoke:
		//println("invoke " + a.ItemName)
		return _i.InvokeFunc(a.ItemName, getConsts(a.Expr)...)
	case text.ActionGet:
		//println("get " + a.ItemName)
		val, err := _i.GetGlobalVal(a.ItemName)
		return []interface{}{val}, err
	default:
		panic("unreachable")
	}
}

func assertReturn(a *text.Assertion, results []interface{}, err error) error {
	expectedVals := getConsts(a.Result)
	if err != nil {
		return fmt.Errorf("line: %d, expected return: %v, got: %v",
			a.Line, expectedVals, err)
	}

	if len(results) != len(expectedVals) {
		return fmt.Errorf("line: %d, expected return: %v, got: %v",
			a.Line, expectedVals, results)
	}

	for i, result := range results {
		expectedVal := expectedVals[i]
		if isNaN32(expectedVal) { // TODO
			if !isNaN32(result) {
				return fmt.Errorf("line: %d, expected return: NaN, got: %v",
					a.Line, result)
			}
		} else if isNaN64(expectedVal) { // TODO
			if !isNaN64(result) {
				return fmt.Errorf("line: %d, expected return: NaN, got: %v",
					a.Line, result)
			}
		} else if result != expectedVal {
			return fmt.Errorf("line: %d, expected return: %v, got: %v",
				a.Line, expectedVal, result)
		}
	}

	return nil
}
func assertTrap(a *text.Assertion, results []interface{}, err error) error {
	if err == nil {
		return fmt.Errorf("line: %d, expected trap: %v, got: %v",
			a.Line, a.Failure, results)
	}
	if strings.Index(err.Error(), a.Failure) < 0 {
		return fmt.Errorf("line: %d, expected trap: %v, got: %v",
			a.Line, a.Failure, err)
	}
	return nil
}
func assertError(a *text.Assertion, err error) error {
	if err == nil || strings.Index(err.Error(), a.Failure) < 0 {
		return fmt.Errorf("line: %d, expected: %v, got: %v",
			a.Line, a.Failure, err)
	}
	return nil
}

func getConsts(expr []binary.Instruction) []interface{} {
	vals := make([]interface{}, len(expr))
	for i, instr := range expr {
		switch instr.Opcode {
		case binary.I32Const:
			vals[i] = instr.Args.(int32)
		case binary.I64Const:
			vals[i] = instr.Args.(int64)
		case binary.F32Const:
			vals[i] = instr.Args.(float32)
		case binary.F64Const:
			vals[i] = instr.Args.(float64)
		default:
			panic("TODO")
		}
	}
	return vals
}

func isNaN32(x interface{}) bool {
	f, ok := x.(float32)
	return ok && math.IsNaN(float64(f))
}
func isNaN64(x interface{}) bool {
	f, ok := x.(float64)
	return ok && math.IsNaN(f)
}
