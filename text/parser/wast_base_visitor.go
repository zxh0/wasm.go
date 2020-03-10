// Code generated from ./text/grammar/WAST.g4 by ANTLR 4.8. DO NOT EDIT.

package parser // WAST

import "github.com/antlr/antlr4/runtime/Go/antlr"

type BaseWASTVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseWASTVisitor) VisitScript(ctx *ScriptContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitCmd(ctx *CmdContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitWastModule(ctx *WastModuleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitAction_(ctx *Action_Context) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitAssertion(ctx *AssertionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitExpected(ctx *ExpectedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitMeta(ctx *MetaContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitModule(ctx *ModuleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitWatModule(ctx *WatModuleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitModuleField(ctx *ModuleFieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitTypeDef(ctx *TypeDefContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitImport_(ctx *Import_Context) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitImportDesc(ctx *ImportDescContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitFunc_(ctx *Func_Context) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitFuncLocal(ctx *FuncLocalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitTable(ctx *TableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitMemory(ctx *MemoryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitGlobal(ctx *GlobalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitExport(ctx *ExportContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitExportDesc(ctx *ExportDescContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitStart(ctx *StartContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitElem(ctx *ElemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitData(ctx *DataContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitEmbeddedIm(ctx *EmbeddedImContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitEmbeddedEx(ctx *EmbeddedExContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitTypeUse(ctx *TypeUseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitFuncVars(ctx *FuncVarsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitValType(ctx *ValTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitBlockType(ctx *BlockTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitGlobalType(ctx *GlobalTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitMemoryType(ctx *MemoryTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitTableType(ctx *TableTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitElemType(ctx *ElemTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitLimits(ctx *LimitsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitFuncType(ctx *FuncTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitParam(ctx *ParamContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitResult(ctx *ResultContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitExpr(ctx *ExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitInstr(ctx *InstrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitFoldedInstr(ctx *FoldedInstrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitBlockInstr(ctx *BlockInstrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitPlainInstr(ctx *PlainInstrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitConstInstr(ctx *ConstInstrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitMemArg(ctx *MemArgContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitNat(ctx *NatContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitValue(ctx *ValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseWASTVisitor) VisitVariable(ctx *VariableContext) interface{} {
	return v.VisitChildren(ctx)
}
