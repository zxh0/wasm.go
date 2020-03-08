package text

import (
	"github.com/zxh0/wasm.go/binary"
)

type codeBuilder struct {
	locals     []binary.Locals
	localNames *symbolTable
	labelNames *symbolTable
	blockDepth int
}

func newCodeBuilder() *codeBuilder {
	return &codeBuilder{
		localNames: newSymbolTable("parameter"),
		labelNames: newSymbolTable("label"),
	}
}

/* params & locals */

func (b *codeBuilder) getLocalIdx(_var string) (int, error) {
	b.localNames.kind = "local"
	return b.localNames.getIdx(_var)
}

func (b *codeBuilder) addParam(name string) error {
	return b.localNames.defineName(name)
}

func (b *codeBuilder) addLocal(name string, t binary.ValType) error {
	b.localNames.kind = "local"
	if err := b.localNames.defineName(name); err != nil {
		return err
	}

	n := len(b.locals)
	if n == 0 || b.locals[n-1].Type != t {
		b.locals = append(b.locals, binary.Locals{N: 1, Type: t})
	} else {
		b.locals[n-1].N++
	}
	return nil
}

/* labels */

func (b *codeBuilder) enterBlock() {
	b.blockDepth++
}
func (b *codeBuilder) exitBlock() {
	b.blockDepth--
}

func (b *codeBuilder) defineLabel(name string) {
	b.labelNames.defineLabel(name, b.blockDepth)
}

func (b *codeBuilder) getBrLabelIdx(_var string) (int, error) {
	if _var[0] != '$' {
		idx := int(parseU32(_var))
		if idx > b.blockDepth {
			return -1, newVerificationError("invalid depth: %d (max %d)",
				idx, b.blockDepth)
		}
		return idx, nil
	}
	depth, err := b.labelNames.getIdx(_var)
	return b.blockDepth - depth, err
}
