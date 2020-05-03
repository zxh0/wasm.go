package jitgoloader

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"unsafe"

	"github.com/dearplain/goloader"

	"github.com/zxh0/wasm.go/aot"
	"github.com/zxh0/wasm.go/binary"
)

type CompiledFunc1 = func(uint64) uint64

func CompileFunc(module binary.Module, code binary.Code,
	ft binary.FuncType, idx int) (CompiledFunc1, error) {

	if len(ft.ParamTypes) != 1 || len(ft.ResultTypes) != 1 || !checkExpr(code.Expr) {
		return nil, fmt.Errorf("can not compile code[%d]", idx)
	}
	s := aot.CompileFunc(module, idx)

	f, err := ioutil.TempFile("", "*.wasm.go")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())

	if _, err = f.WriteString(s); err != nil {
		return nil, err
	}

	objFilename := f.Name() + ".o"
	defer os.Remove(objFilename)

	cmd := exec.Command("go", "tool", "compile", "-o", objFilename, f.Name())
	if err = cmd.Run(); err != nil {
		return nil, err
	}

	f1, err := loadObj(objFilename)
	if err != nil {
		return nil, err
	}

	return f1, nil
}

func checkExpr(expr binary.Expr) bool {
	for _, instr := range expr {
		op := instr.Opcode
		if (op == binary.Call || op == binary.CallIndirect) ||
			(op == binary.GlobalGet || op == binary.GlobalSet) ||
			(op >= binary.I32Load && op <= binary.MemoryGrow) {
			return false
		}
		if op == binary.Block || op == binary.Loop {
			if !checkExpr(instr.Args.(binary.BlockArgs).Instrs) {
				return false
			}
		}
		if op == binary.If {
			ifArgs := instr.Args.(binary.IfArgs)
			if !checkExpr(ifArgs.Instrs1) || !checkExpr(ifArgs.Instrs2) {
				return false
			}
		}
	}
	return true
}

func loadObj(objFilename string) (CompiledFunc1, error) {
	cr, err := goloader.ReadObjs([]string{objFilename}, []string{""})
	if err != nil {
		return nil, err
	}

	symPtr := make(map[string]uintptr)
	goloader.RegSymbol(symPtr)

	cm, err := goloader.Load(cr, symPtr)
	if err != nil {
		return nil, err
	}

	runFuncPtr := cm.Syms["main.Call"]
	funcPtrContainer := (uintptr)(unsafe.Pointer(&runFuncPtr))
	cf := *(*CompiledFunc1)(unsafe.Pointer(&funcPtrContainer))
	return cf, nil
}
