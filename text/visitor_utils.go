package text

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/text/parser"
)

func getExpr(node parser.IExprContext, v antlr.ParseTreeVisitor) []binary.Instruction {
	if node == nil {
		return []binary.Instruction{}
	} else {
		return node.Accept(v).([]binary.Instruction)
	}
}

func getText(node interface{ GetText() string }) string {
	if node == nil {
		return ""
	}
	return node.GetText()
}

func getStr(node antlr.TerminalNode) string {
	text := node.GetText()
	return text[1 : len(text)-1]
}
func getAllStr(nodes []antlr.TerminalNode) string {
	s := ""
	for _, node := range nodes {
		s += getStr(node)
	}
	return s
}
