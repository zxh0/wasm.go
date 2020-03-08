package text

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/zxh0/wasm.go/text/parser"
)

type errorReporter struct {
	reportsValidationError bool // TODO: rename
}

func (reporter errorReporter) reportErr(err error, node interface{}) {
	if err != nil {
		switch x := err.(type) {
		case *SemanticError:
			x.token = getToken(node)
			panic(err)
		case *ValidationError:
			x.token = getToken(node)
			if reporter.reportsValidationError {
				panic(err)
			}
		default:
			panic(err) // TODO
		}
	}
}

func getToken(node interface{}) antlr.Token {
	switch x := node.(type) {
	case antlr.Token:
		return x
	case antlr.TerminalNode:
		return x.GetSymbol()
	case parser.IVariableContext:
		return getToken(x.GetChild(0))
	default:
		panic("TODO")
	}
}
