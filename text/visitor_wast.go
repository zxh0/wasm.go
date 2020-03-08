package text

import (
	"math"

	"github.com/zxh0/wasm.go/binary"
	"github.com/zxh0/wasm.go/text/parser"
)

var _ parser.WASTVisitor = (*wastVisitor)(nil)

type wastVisitor struct {
	watVisitor
	scriptBuilder *scriptBuilder
}

func newWastVisitor() *wastVisitor {
	return &wastVisitor{}
}

func (v *wastVisitor) VisitScript(ctx *parser.ScriptContext) interface{} {
	v.scriptBuilder = newScriptBuilder()
	for _, cmd := range ctx.AllCmd() {
		v.scriptBuilder.addCmd(cmd.Accept(v))
	}
	return v.scriptBuilder.script
}

func (v *wastVisitor) VisitCmd(ctx *parser.CmdContext) interface{} {
	if ctx.WastModule() != nil {
		return ctx.WastModule().Accept(v)
	} else if ctx.Action_() != nil {
		return ctx.Action_().Accept(v)
	} else if ctx.Assertion() != nil {
		return ctx.Assertion().Accept(v)
	} else if ctx.Meta() != nil {
		return ctx.Meta().Accept(v)
	} else { // register
		return &Register{
			ModuleName: getStr(ctx.STRING()),
			Name:       getText(ctx.NAME()),
		}
	}
}

func (v *wastVisitor) VisitWastModule(ctx *parser.WastModuleContext) interface{} {
	if ctx.WatModule() != nil {
		return ctx.WatModule().Accept(v)
	}

	name := getText(ctx.NAME())
	switch ctx.GetKind().GetText() {
	case "binary":
		return &BinaryModule{
			Name: name,
			Data: escape(getAllStr(ctx.AllSTRING())),
		}
	case "quote":
		return &QuotedModule{
			Name: name,
			Text: getAllStr(ctx.AllSTRING()),
		}
	default:
		panic("unreachable")
	}
}

func (v *wastVisitor) VisitAction_(ctx *parser.Action_Context) interface{} {
	a := Action{}

	switch ctx.GetKind().GetText() {
	case "invoke":
		a.Kind = ActionInvoke
	case "get":
		a.Kind = ActionGet
	default:
		panic("unreachable")
	}

	a.ModuleName = getText(ctx.NAME())
	a.ItemName = getStr(ctx.STRING())
	if a.Kind == ActionInvoke {
		a.Expr = ctx.Expr().Accept(v).([]binary.Instruction)
	}

	return &a
}

func (v *wastVisitor) VisitAssertion(ctx *parser.AssertionContext) interface{} {
	a := Assertion{}

	switch ctx.GetKind().GetText() {
	case "assert_return":
		a.Kind = AssertReturn
	case "assert_trap":
		a.Kind = AssertTrap
	case "assert_exhaustion":
		a.Kind = AssertExhaustion
	case "assert_malformed":
		a.Kind = AssertMalformed
	case "assert_invalid":
		a.Kind = AssertInvalid
	case "assert_unlinkable":
		a.Kind = AssertUnlinkable
	default:
		panic("unreachable")
	}

	if ctx.Action_() != nil {
		a.Action = ctx.Action_().Accept(v).(*Action)
	}
	for _, result := range ctx.AllExpected() {
		a.Result = append(a.Result, result.Accept(v).(binary.Instruction))
	}
	if ctx.WastModule() != nil {
		a.Module = ctx.WastModule().Accept(v)
	}
	if ctx.STRING() != nil {
		a.Failure = getStr(ctx.STRING())
	}

	return &a
}
func (v *wastVisitor) VisitExpected(ctx *parser.ExpectedContext) interface{} {
	if ctx.GetNan() != nil {
		instr := newInstruction(ctx.GetOp().GetText())
		switch instr.Opcode {
		case binary.F32Const:
			instr.Args = float32(math.NaN())
		case binary.F64Const:
			instr.Args = math.NaN()
		default:
			panic("TODO:NaN")
		}
		return instr
	}
	return ctx.ConstInstr().Accept(v).(binary.Instruction)
}

func (v *wastVisitor) VisitMeta(ctx *parser.MetaContext) interface{} {
	panic("implement me")
}
