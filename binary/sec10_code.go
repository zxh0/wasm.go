package binary

import "fmt"

//type CodeSec = []Code

type Code struct {
	Locals []Locals
	Expr   Expr
}

type Locals struct {
	N    uint32
	Type ValType
}

func readCodeSec(reader *WasmReader) (vec []Code, err error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	vec = make([]Code, n)
	for i := range vec {
		if vec[i], err = readCode(reader); err != nil {
			return
		}
	}
	return
}

func readCode(reader *WasmReader) (code Code, err error) {
	var contents []byte
	if contents, err = reader.readBytes(); err != nil {
		return
	}

	codeReader := WasmReader{data: contents}
	if code.Locals, err = readLocalsVec(&codeReader); err != nil {
		return
	}
	if code.Expr, err = readExpr(&codeReader); err != nil {
		return
	}
	if codeReader.remaining() > 0 {
		err = fmt.Errorf("invalid code") // TODO
	}

	return
}

func readLocalsVec(reader *WasmReader) (vec []Locals, err error) {
	n, err := reader.readVarU32()
	if err != nil {
		return nil, err
	}

	vec = make([]Locals, n)
	for i := range vec {
		if vec[i], err = readLocals(reader); err != nil {
			return
		}
	}
	return
}

func readLocals(reader *WasmReader) (locals Locals, err error) {
	if locals.N, err = reader.readVarU32(); err != nil {
		return
	}
	locals.Type, err = readValType(reader)
	return
}

func (code Code) GetLocalCount() int {
	n := 0
	for _, locals := range code.Locals {
		n += int(locals.N)
	}
	return n
}
