package text

import "github.com/zxh0/wasm.go/binary"

const (
	ActionInvoke = 1
	ActionGet    = 2
)

const (
	AssertReturn     = 1
	AssertTrap       = 2
	AssertExhaustion = 3
	AssertMalformed  = 4
	AssertInvalid    = 5
	AssertUnlinkable = 6
)

// https://github.com/WebAssembly/spec/tree/master/interpreter#scripts
type Script struct {
	Cmds []interface{}
}

type WatModule struct {
	Line   int
	Name   string
	Module *binary.Module
}
type BinaryModule struct {
	Line int
	Name string
	Data []byte
}
type QuotedModule struct {
	Line int
	Name string
	Text string
}

type Register struct {
	ModuleName string
	Name       string
}

type Action struct {
	Kind       byte
	ModuleName string
	ItemName   string
	Expr       []binary.Instruction
}

type Assertion struct {
	Line    int
	Kind    byte
	Action  *Action
	Result  []binary.Instruction
	Module  interface{}
	Failure string
}

type Meta struct {
	// TODO
}
