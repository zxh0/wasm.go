// Code generated from ./text/grammar/WAST.g4 by ANTLR 4.8. DO NOT EDIT.

package parser // WAST

import "github.com/antlr/antlr4/runtime/Go/antlr"

// A complete Visitor for a parse tree produced by WASTParser.
type WASTVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by WASTParser#script.
	VisitScript(ctx *ScriptContext) interface{}

	// Visit a parse tree produced by WASTParser#cmd.
	VisitCmd(ctx *CmdContext) interface{}

	// Visit a parse tree produced by WASTParser#wastModule.
	VisitWastModule(ctx *WastModuleContext) interface{}

	// Visit a parse tree produced by WASTParser#action_.
	VisitAction_(ctx *Action_Context) interface{}

	// Visit a parse tree produced by WASTParser#assertion.
	VisitAssertion(ctx *AssertionContext) interface{}

	// Visit a parse tree produced by WASTParser#expected.
	VisitExpected(ctx *ExpectedContext) interface{}

	// Visit a parse tree produced by WASTParser#meta.
	VisitMeta(ctx *MetaContext) interface{}

	// Visit a parse tree produced by WASTParser#module.
	VisitModule(ctx *ModuleContext) interface{}

	// Visit a parse tree produced by WASTParser#watModule.
	VisitWatModule(ctx *WatModuleContext) interface{}

	// Visit a parse tree produced by WASTParser#moduleField.
	VisitModuleField(ctx *ModuleFieldContext) interface{}

	// Visit a parse tree produced by WASTParser#typeDef.
	VisitTypeDef(ctx *TypeDefContext) interface{}

	// Visit a parse tree produced by WASTParser#import_.
	VisitImport_(ctx *Import_Context) interface{}

	// Visit a parse tree produced by WASTParser#importDesc.
	VisitImportDesc(ctx *ImportDescContext) interface{}

	// Visit a parse tree produced by WASTParser#func_.
	VisitFunc_(ctx *Func_Context) interface{}

	// Visit a parse tree produced by WASTParser#funcLocal.
	VisitFuncLocal(ctx *FuncLocalContext) interface{}

	// Visit a parse tree produced by WASTParser#table.
	VisitTable(ctx *TableContext) interface{}

	// Visit a parse tree produced by WASTParser#memory.
	VisitMemory(ctx *MemoryContext) interface{}

	// Visit a parse tree produced by WASTParser#global.
	VisitGlobal(ctx *GlobalContext) interface{}

	// Visit a parse tree produced by WASTParser#export.
	VisitExport(ctx *ExportContext) interface{}

	// Visit a parse tree produced by WASTParser#exportDesc.
	VisitExportDesc(ctx *ExportDescContext) interface{}

	// Visit a parse tree produced by WASTParser#start.
	VisitStart(ctx *StartContext) interface{}

	// Visit a parse tree produced by WASTParser#elem.
	VisitElem(ctx *ElemContext) interface{}

	// Visit a parse tree produced by WASTParser#data.
	VisitData(ctx *DataContext) interface{}

	// Visit a parse tree produced by WASTParser#embeddedIm.
	VisitEmbeddedIm(ctx *EmbeddedImContext) interface{}

	// Visit a parse tree produced by WASTParser#embeddedEx.
	VisitEmbeddedEx(ctx *EmbeddedExContext) interface{}

	// Visit a parse tree produced by WASTParser#typeUse.
	VisitTypeUse(ctx *TypeUseContext) interface{}

	// Visit a parse tree produced by WASTParser#funcVars.
	VisitFuncVars(ctx *FuncVarsContext) interface{}

	// Visit a parse tree produced by WASTParser#valType.
	VisitValType(ctx *ValTypeContext) interface{}

	// Visit a parse tree produced by WASTParser#blockType.
	VisitBlockType(ctx *BlockTypeContext) interface{}

	// Visit a parse tree produced by WASTParser#globalType.
	VisitGlobalType(ctx *GlobalTypeContext) interface{}

	// Visit a parse tree produced by WASTParser#memoryType.
	VisitMemoryType(ctx *MemoryTypeContext) interface{}

	// Visit a parse tree produced by WASTParser#tableType.
	VisitTableType(ctx *TableTypeContext) interface{}

	// Visit a parse tree produced by WASTParser#elemType.
	VisitElemType(ctx *ElemTypeContext) interface{}

	// Visit a parse tree produced by WASTParser#limits.
	VisitLimits(ctx *LimitsContext) interface{}

	// Visit a parse tree produced by WASTParser#funcType.
	VisitFuncType(ctx *FuncTypeContext) interface{}

	// Visit a parse tree produced by WASTParser#param.
	VisitParam(ctx *ParamContext) interface{}

	// Visit a parse tree produced by WASTParser#result.
	VisitResult(ctx *ResultContext) interface{}

	// Visit a parse tree produced by WASTParser#expr.
	VisitExpr(ctx *ExprContext) interface{}

	// Visit a parse tree produced by WASTParser#instr.
	VisitInstr(ctx *InstrContext) interface{}

	// Visit a parse tree produced by WASTParser#foldedInstr.
	VisitFoldedInstr(ctx *FoldedInstrContext) interface{}

	// Visit a parse tree produced by WASTParser#blockInstr.
	VisitBlockInstr(ctx *BlockInstrContext) interface{}

	// Visit a parse tree produced by WASTParser#plainInstr.
	VisitPlainInstr(ctx *PlainInstrContext) interface{}

	// Visit a parse tree produced by WASTParser#constInstr.
	VisitConstInstr(ctx *ConstInstrContext) interface{}

	// Visit a parse tree produced by WASTParser#memArg.
	VisitMemArg(ctx *MemArgContext) interface{}

	// Visit a parse tree produced by WASTParser#nat.
	VisitNat(ctx *NatContext) interface{}

	// Visit a parse tree produced by WASTParser#value.
	VisitValue(ctx *ValueContext) interface{}

	// Visit a parse tree produced by WASTParser#variable.
	VisitVariable(ctx *VariableContext) interface{}
}
