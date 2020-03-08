package text

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/zxh0/wasm.go/text/parser"
)

var _ parser.WASTVisitor = (*watNamesVisitor)(nil)

// build symbol table
type watNamesVisitor struct {
	baseVisitor
	errorReporter
	moduleBuilder *moduleBuilder
}

func (v *watNamesVisitor) VisitModuleField(ctx *parser.ModuleFieldContext) interface{} {
	if imp := ctx.Import_(); imp != nil {
		return imp.Accept(v)
	}
	if f := ctx.Func_(); f != nil {
		return f.Accept(v)
	}
	if t := ctx.Table(); t != nil {
		return t.Accept(v)
	}
	if m := ctx.Memory(); m != nil {
		return m.Accept(v)
	}
	if g := ctx.Global(); g != nil {
		return g.Accept(v)
	}
	return nil
}

func (v *watNamesVisitor) VisitImport_(ctx *parser.Import_Context) interface{} {
	if err := v.moduleBuilder.ensureNoNonImports(); err != nil {
		v.reportErr(err, ctx.GetChild(1))
	}
	return ctx.ImportDesc().Accept(v)
}
func (v *watNamesVisitor) VisitImportDesc(ctx *parser.ImportDescContext) interface{} {
	kind := ctx.GetKind().GetText()
	name := getText(ctx.NAME())
	if err := v.moduleBuilder.checkCount(kind); err != nil {
		v.reportErr(err, ctx.GetParent().GetChild(1))
	}
	if err := v.moduleBuilder.importName(kind, name); err != nil {
		v.reportErr(err, ctx.NAME())
	}
	return nil
}

func (v *watNamesVisitor) VisitFunc_(ctx *parser.Func_Context) interface{} {
	return v.visitModuleField(ctx, "func", ctx.NAME(), ctx.EmbeddedIm())
}
func (v *watNamesVisitor) VisitTable(ctx *parser.TableContext) interface{} {
	return v.visitModuleField(ctx, "table", ctx.NAME(), ctx.EmbeddedIm())
}
func (v *watNamesVisitor) VisitMemory(ctx *parser.MemoryContext) interface{} {
	return v.visitModuleField(ctx, "memory", ctx.NAME(), ctx.EmbeddedIm())
}
func (v *watNamesVisitor) VisitGlobal(ctx *parser.GlobalContext) interface{} {
	return v.visitModuleField(ctx, "global", ctx.NAME(), ctx.EmbeddedIm())
}

func (v *watNamesVisitor) visitModuleField(ctx antlr.Tree, kind string,
	name antlr.TerminalNode, embeddedIm parser.IEmbeddedImContext) interface{} {

	if err := v.moduleBuilder.checkCount(kind); err != nil {
		v.reportErr(err, ctx.GetChild(1))
	}
	if embeddedIm != nil {
		if err := v.moduleBuilder.ensureNoNonImports(); err != nil {
			v.reportErr(err, embeddedIm.GetChild(1))
		}
		if err := v.moduleBuilder.importName(kind, getText(name)); err != nil {
			v.reportErr(err, name)
		}
	} else {
		if err := v.moduleBuilder.defineName(kind, getText(name)); err != nil {
			v.reportErr(err, name)
		}
	}
	return nil
}
