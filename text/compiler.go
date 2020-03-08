package text

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/text/parser"
)

// WAT Module
func CompileModuleFile(filename string) (*binary.Module, error) {
	input, err := antlr.NewFileStream(filename)
	if err != nil {
		return nil, err
	}
	return CompileModule(input)
}
func CompileModuleStr(s string) (*binary.Module, error) {
	input := antlr.NewInputStream(s)
	return CompileModule(input)
}
func CompileModule(input antlr.CharStream) (m *binary.Module, err error) {
	errListener := &ErrorListener{}
	p := newParser(input, errListener)
	ctx := p.Module()
	if err = errListener.GetErrors(input); err != nil {
		return
	}
	defer func() {
		if _err := recover(); _err != nil {
			err = fillDetail(_err, input)
		}
	}()
	m = ctx.Accept(newWatVisitor()).(*binary.Module)
	return
}

// WAST Script
func CompileScriptFile(filename string) (*Script, error) {
	input, err := antlr.NewFileStream(filename)
	if err != nil {
		return nil, err
	}
	return CompileScript(input)
}
func CompileScriptStr(s string) (*Script, error) {
	input := antlr.NewInputStream(s)
	return CompileScript(input)
}
func CompileScript(input antlr.CharStream) (s *Script, err error) {
	errListener := &ErrorListener{}
	p := newParser(input, errListener)
	ctx := p.Script()
	if err = errListener.GetErrors(input); err != nil {
		return
	}
	defer func() {
		if _err := recover(); _err != nil {
			err = fillDetail(_err, input)
		}
	}()
	s = ctx.Accept(newWastVisitor()).(*Script)
	return
}

func newParser(input antlr.CharStream,
	errListener antlr.ErrorListener) *parser.WASTParser {

	lexer := parser.NewWASTLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewWASTParser(stream)
	p.RemoveErrorListeners()
	p.AddErrorListener(errListener)
	//p.AddErrorListener(antlr.NewDiagnosticErrorListener(true))
	p.BuildParseTrees = true
	return p
}

func fillDetail(err interface{}, input antlr.CharStream) error {
	switch x := err.(type) {
	case *SemanticError:
		x.FillDetail(input)
		return x
	case *ValidationError:
		x.FillDetail(input)
		return x
	default:
		panic(err) // TODO
	}
}
